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
