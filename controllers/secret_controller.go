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
	"fmt"
	"time" // Needed for requeue

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
	"math/big"

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
		logger.Info("Secret will have name " + secret.Spec.Secret.Name + " and namespace " + secret.Spec.Secret.Namespace + " and type " + secret.Spec.Generator.Type + " with length " + fmt.Sprintf("%v", secret.Spec.Generator.Length))
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

	var randomString []byte

	secretData := make(map[string][]byte)

	var charset string

	if secret.Spec.Generator.Charset != "" {
		charset = secret.Spec.Generator.Charset
	} else {
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	switch secret.Spec.Generator.Type {
	case "authelia-hash":
		randomToBeHashedString := randomStringGenerator(secret.Spec.Generator.Length, charset)

		logger.Info("Hashed String is: " + string(randomToBeHashedString))

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

		hashedString := digest.Encode()

		logger.Info("Key is: " + hashedString)

		secretData := make(map[string][]byte)

		if secret.Spec.Generator.HashKey == "" {
			secret.Spec.Generator.HashKey = secret.Spec.Generator.Key + "_HASHED" // If the HashKey doesn't get defined, generate a default one
		}

		secretData[secret.Spec.Generator.HashKey] = []byte(hashedString)
		secretData[secret.Spec.Generator.Key] = randomToBeHashedString

	case "string":
		randomString = randomStringGenerator(secret.Spec.Generator.Length, charset)
		secretData := make(map[string][]byte)
		secretData[secret.Spec.Generator.Key] = randomString

	default:
		logger.Error(nil, "No valid generator given")
		return nil
	}
	secretDefinition := corev1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Spec.Secret.Name,
			Namespace: secret.Spec.Secret.Namespace,
		},
		Data: secretData,
	}
	ctrl.SetControllerReference(secret, &secretDefinition, r.Scheme)

	return &secretDefinition

}

// SetupWithManager sets up the controller with the Manager.
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretsv1alpha1.Secret{}).
		Complete(r)
}

func randomStringGenerator(length int, charset string) []byte {
	charsetLength := big.NewInt(int64(len(charset)))

	randomString := make([]byte, length)
	for i := range randomString {
		randomIndex, _ := rand.Int(rand.Reader, charsetLength)
		randomString[i] = charset[randomIndex.Int64()]
	}

	return randomString
}
