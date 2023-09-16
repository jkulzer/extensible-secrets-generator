# extensible-secrets-generator

## Description
A simple Kubernetes operator to generate secrets with random values

## Getting Started

### Running on the cluster

TODO: Container image and Helm Chart doesn't exist yet

#### Development instance

You can run the operator locally on your machine for development with this command:

```
make install run
```

## Generating a secret

A example CRD

This creates a secret with the following metadata:

* The name `name`
* The namespace `default`
* two keys named `key` and `hashKey`
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
  generator:
    type: authelia-hash
    length: 10
    key: testkey
    hashKey: hashed
```

### Possible CRD options


#### Secret info
```
spec:
  secret:
    name: name
    namespace: default
```
#### Generator type

Specifies the kind of secret that should be generated. Currently the two options are
* string
* authelia-hash
```
---
spec:
  generator:
    type: string | authelia-hash 
```

##### String
```
---
spec:
  generator:
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
  generator:
    type: authelia-hash
    length: 10
    key: testkey
    hashKey: hashed
```
The `length` key specifies how long the randomly generated cleartext secret should be

The `charset` key specifies what characters the string should contain. Defaults to `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`

the `key` key specifies under which key the random plaintext string should be stored in the Kubernetes secret

the `hashKey` key specifies under which key the hashed version of the random plaintext string should be stored in the Kubernetes secret. If not set, it default to the `key` key + `_HASHED`
