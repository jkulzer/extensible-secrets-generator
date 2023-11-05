/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"time" // Needed for requeue

	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/client-go/applyconfigurations/admissionregistration/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	// v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// read error messages from the K8s api
	"k8s.io/apimachinery/pkg/api/errors"

	// for random secrets generation
	"crypto/rand"
	"encoding/base64"

	// for hashing
	"github.com/go-crypt/crypt/algorithm"
	"github.com/go-crypt/crypt/algorithm/pbkdf2"

	secretsv1alpha1 "github.com/jkulzer/extensible-secrets-generator/api/v1alpha1"
)

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=secrets.esg.jkulzer.dev,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=secrets.esg.jkulzer.dev,resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO: user: Modify the Reconcile function to compare the state specified by
// the Secret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Starting reconcile loop")
	secret := &secretsv1alpha1.Secret{}
	err := r.Get(ctx, req.NamespacedName, secret)

	found := corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: secret.Spec.Secret.Name, Namespace: secret.Spec.Secret.Namespace}, &found)

	if err != nil && errors.IsNotFound(err) {
		// Define a new secret
		newSecret := r.secretGeneration(secret, ctx)
		logger.Info("Secret will have name " + secret.Spec.Secret.Name + " and namespace " + secret.Spec.Secret.Namespace)
		logger.Info("Creating Secret" + "Name: " + secret.Spec.Secret.Name + "Namespace: " + secret.Spec.Secret.Namespace)
		if err = r.Create(ctx, newSecret); err != nil {
			logger.Error(err, "Failed to create new Secret",
				"Secret.Name", secret.Spec.Secret.Name, "Secret.Namespace", secret.Spec.Secret.Namespace)
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: time.Minute}, nil

	} else if err != nil {
		logger.Error(err, "Failed to get Secrets")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, err
}

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
func (r *SecretReconciler) secretGeneration(secret *secretsv1alpha1.Secret, ctx context.Context) *corev1.Secret {
	logger := log.FromContext(ctx)

	logger.Info("Creating Secret with Name " + secret.Spec.Secret.Name + " and Namespace " + secret.Spec.Secret.Namespace)

	secretData := make(map[string][]byte)

	keys := secret.Spec.Keys

	generators := secret.Spec.Generators

	generatedSecrets := make(map[string]string)

	for _, generator := range generators {
		secret1, secret2 := generateSecret(generator, ctx)
		if generator.Type == "authelia-hash" {
			generatedSecrets[generator.Name+".hashed"] = secret2
		}
		generatedSecrets[generator.Name] = secret1
	}

	for _, key := range keys {
		for mapKey, generatedSecret := range generatedSecrets {
			logger.Info(mapKey)
			var templateString string
			if key.TemplateString == string(secretData[key.Key]) || string(secretData[key.Key]) == "" {
				templateString = key.TemplateString
			} else {
				templateString = string(secretData[key.Key])
			}
			secretData[key.Key] = []byte(strings.ReplaceAll(templateString, "{{ "+mapKey+" }}", generatedSecret))
		}
	}

	secretDefinition := corev1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Spec.Secret.Name,
			Namespace: secret.Spec.Secret.Namespace,
			Labels:    secret.Spec.Secret.Labels,
		},
		Data: secretData,
	}
	ctrl.SetControllerReference(secret, &secretDefinition, r.Scheme)

	return &secretDefinition

}

func generateSecret(generator secretsv1alpha1.SecretGenerator, ctx context.Context) (string, string) {

	logger := log.FromContext(ctx)

	var charset string

	if generator.Charset != "" {
		charset = generator.Charset
	} else {
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	secret1 := ""
	secret2 := ""

	switch generator.Type {
	case "authelia-hash":
		logger := log.FromContext(ctx)

		randomToBeHashedString := randomStringGenerator(generator.Length, generator.Charset)

		var (
			hasher *pbkdf2.Hasher
			err    error
			digest algorithm.Digest
		)

		if hasher, err = pbkdf2.New(
			pbkdf2.WithVariantName("sha512"),
			pbkdf2.WithIterations(310000),
			pbkdf2.WithSaltLength(16),
		); err != nil {
			logger.Error(err, "Failed to create hasher")
		}

		if digest, err = hasher.Hash(string(randomToBeHashedString)); err != nil {
			logger.Error(err, "Failed to hash string")
		}

		secret1 = string(randomToBeHashedString)

		secret2 = digest.Encode()

	case "string":
		randomString := randomStringGenerator(generator.Length, charset)
		secret1 = string(randomString)
	default:
		logger.Error(nil, "No valid generator given")
		return "", ""
	}

	return secret1, secret2
}

// SetupWithManager sets up the controller with the Manager.
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretsv1alpha1.Secret{}).
		Complete(r)
}

func randomStringGenerator(length int, charset string) []byte {

	randomBytes := make([]byte, length)

	// Read cryptographically secure random bytes into the slice
	rand.Read(randomBytes)

	// Encode the random bytes to a base64 string
	randomString := base64.StdEncoding.EncodeToString(randomBytes)

	// Trim any padding characters and select the desired length
	randomString = randomString[:length]

	return []byte(randomString)
}
