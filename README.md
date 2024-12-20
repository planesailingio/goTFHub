# goTFHub - Overview
`goTFHub` is a stateless service that is built to host Terraform providers, modules and state files.
This is particularly useful if you're using Terraform-based tooling in internet-disconnected environments e.g. Terraform, OpenTofu, Coder etc.

The service is statless and backed by an S3 object store - Any S3 API-compatible object store, such as MinIO can be used.


## Getting Started
goTFHub is designed to work in Kubernetes and so there is a Helm chart for easy deployment.

```
helm repo add https://charts.planesailing.io
```



### Artefact Manifest



### Environment Variables

| Variable | Default | Example | Description |
| :------- | :------ | ------- | ----------- |
|          |         |         |             |

## Commands


### Pull

Command: `gotfhub pull`

| Parameter | Default | Description |
| :-------- | :------ | ----------- |
|  |         |             |


### Kubernetes-based Deployment
Refer to the helm chart for more information.

### Seamless Experience
Add `registry.terraform.io` to your internal DNS and ensure it resolves to the Kubernetes Ingress to provide a seamless experience to your engineers.

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
