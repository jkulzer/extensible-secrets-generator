# extensible-secrets-generator

## Description
A simple Kubernetes operator to generate secrets with random values

## Getting Started

### Running on the cluster
1. Add the Helm repo
```
helm repo add extensible-secrets-generator https://jkulzer.github.io/extensible-secrets-generator
```
2. (Optional) Install with default values
```
helm install extensible-secrets-generator extensible-secrets-generator/extensible-secrets-generator
```

### Development instance

You can run the operator locally on your machine for development with this command:

```
make install run
```

## Generating a secret

A example CRD

This creates a secret with the following metadata:

* The name `name`
* The namespace `default`
* two keys named `KEY` and `HASH_KEY`
* it generates the keys using the authelia-hash generator
* the length of the key is 10

```
---
apiVersion: secrets.esg.jkulzer.dev/v1alpha1
kind: Secret
metadata:
  name: test-secret
  namespace: default
spec:
  secret:
    name: name
    namespace: default
  keys:
    - key: KEY
      templateString: "{{ TEST }}"
    - key: HASH_KEY
      templateString: "{{ TEST.hashed }}"
  generators:
     - name: TEST
       type: authelia-hash
       length: 10
```

### Possible CRD options


#### Secret info
```
spec:
  secret:
    name: name
    namespace: default
    labels:
      label-that-should-be-present-on-the-secret.k8s.io: true
```

The `spec.secret.labels` field will get added to the secret that gets generated

#### Generator type

Specifies the kind of secret that should be generated. Currently the two options are
* string
* authelia-hash
```
---
spec:
  generators:
   - name: ...
     type: string | authelia-hash 
```

##### String
```
---
spec:
  generators:
   - name: ...
     type: string
     length: 20
     charset: abcdefghijklmnopqrstuvwxyz
```

The `length` key specifies how long the randomly generated secret should be

The `charset` key specifies what characters the string should contain. Defaults to `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`

the `key` key specifies under which key the random string should be stored in the Kubernetes secret

##### authelia-hash

```
---
apiVersion: secrets.esg.jkulzer.dev/v1alpha1
kind: Secret
metadata:
  name: test-secret
  namespace: default
spec:
  secret:
    name: name
    namespace: default
  keys:
  - name: MY_KEY
    templateString: "{{ TEST }}"
  - name: MY_HASHED_KEY
    templateString: "{{ TEST.hashed }}"
  generators:
  - name: TEST
    type: authelia-hash
    length: 10
```
The `length` key specifies how long the randomly generated cleartext secret should be
Using `GENERATOR_NAME.hashed` you can access the Authelia-compatible hashed version of the string
