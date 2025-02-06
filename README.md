# goTFHub - Overview
Project Status: BETA
`goTFHub` is a stateless service that is built to host Terraform providers, modules and state files.
This is particularly useful if you're using Terraform-based tooling in internet-disconnected environments e.g. Terraform, OpenTofu, Coder etc.

The service is statless and backed by an S3 object store - Any S3 API-compatible object store, such as MinIO can be used.

The tool is also a single binary that can be used to mirror desired Terraform Providers & Modules.


- [goTFHub - Overview](#gotfhub---overview)
  - [Getting Started](#getting-started)
    - [Artefact Manifest](#artefact-manifest)
  - [Commands](#commands)
    - [Synchronise Providers \& Modules (Pull)](#synchronise-providers--modules-pull)
      - [Example `artefacts.yaml`](#example-artefactsyaml)
    - [Synchronise Providers \& Modules - Push](#synchronise-providers--modules---push)
    - [Kubernetes-based Deployment](#kubernetes-based-deployment)
      - [Configuration Example](#configuration-example)
        - [AWS](#aws)
          - [IRSA](#irsa)
          - [Static Credentials](#static-credentials)
        - [MinIO](#minio)
    - [Seamless Experience](#seamless-experience)
  - [Example Terraform Configuration](#example-terraform-configuration)


## Getting Started
goTFHub is designed to work in Kubernetes and so there is a Helm chart for easy deployment.

```
helm repo add planesailingio https://charts.planesailing.io
```



### Artefact Manifest

## Commands

### Synchronise Providers & Modules (Pull)
The `gotfhub pull` command downloads the desired artefacts to then be able to mirror to your offline environment.

Command: `gotfhub pull --artefact-path ./artefacts.yaml`


#### Example `artefacts.yaml`
The `artefacts.yaml` is used to define which providers and module versions to mirror.
An example is:
```
providers:
  - namespace: coder/coder
  - namespace: coder/coderd
  - namespace: cyrilgdn/postgresql
  - namespace: cyrilgdn/rabbitmq
  - namespace: gitlabhq/gitlab
    version: 3.9.1
  - namespace: gitlabhq/gitlab
  - namespace: hashicorp/archive
modules:
  - namespace: terraform-aws-modules/s3-bucket/aws
    version: 4.2.2
```

### Synchronise Providers & Modules - Push
Pushing to the S3 bucket is simple and can be done manually or via the tool.
Command: `gotfhub push --backend aws --bucket my_s3_bucket --local-path ./terraform-registry`

> NOTE: You must have the AWS or MINIO credentials already configured. If using MINIO, ensure the relevant environment variables are configured.


### Kubernetes-based Deployment
Refer to the helm chart values file for more information.
```
helm 
```
#### Configuration Example
You must supply the relevant environment variables

##### AWS
###### IRSA
You can leverage IRSA and mount the identity from the `ServiceAccount`.     
You must set `s3.credentials.useStaticCreds` to `false`

###### Static Credentials
Here's an example for AWS using static credentials.

```
apiVersion: v1
kind: Secret
metadata:
  name:  s3-credentials
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: 08ca1907-810f-48d5-b325-9addcd4f82c5
  AWS_SECRET_ACCESS_KEY: 08ca1907-810f-48d5-b325-9addcd4f82c5
  AWS_REGION: us-east-1
```

##### MinIO
Here's an example for MinIO using static credentials. By default the Helm chart is configured to use the secret `s3-credentials`. You can update the reference in the values file `s3.credentials.existingSecret`.

```
apiVersion: v1
kind: Secret
metadata:
  name:  s3-credentials
type: Opaque
stringData:
  MINIO_ENDPOINT: http://miniohost-api:9000
  MINIO_ACCESS_KEY_ID: 08ca1907-810f-48d5-b325-9addcd4f82c5
  MINIO_SECRET_ACCESS_KEY: 08ca1907-810f-48d5-b325-9addcd4f82c5
  MINIO_REGION: us-east-1
  MINIO_TLS_ENABLE: true
  MINIO_TLS_SKIP_VERIFY: false
```




### Seamless Experience
Add `registry.terraform.io` to your DNS and ensure it resolves to the Kubernetes Ingress to provide a seamless experience to your consumers.

## Example Terraform Configuration
```
terraform {
  required_providers {
    kubernetes = {
      source = "terraform-registry.ngrok-free.app/coder/coder"
    }
  }
}

module "myiamuser" {
  source = "terraform-registry.ngrok-free.app/terraform-aws-modules/iam/aws"
  version = "4.2.2"
}
```

<!-- ### Environment Variables (Optional)
#### DNS
If you want to provide your users with a more seamless experience when working and defining provider sources', simply set `publishPublicDomains` to true and ensure the following domains resolve to the service domain.

- registry.terraform.io -->
