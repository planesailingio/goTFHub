# goTFHub
`goTFHub` is a stateless service that is built to host Terraform providers, modules and state files.
This is particularly useful if you're using Terraform-based tooling in internet-disconnected environments e.g. Terraform, OpenTofu, Coder etc.

## Architecture
The service is
1. written in Go
2. Exposes several HTTP REST endpoints
3. Backed by S3 object store - In practice, any S3 API-compatible object store, such as MinIO can be used

## Installation

### Kubernetes-based Deployment
Refer to the helm chart for more information.

### Docker (Compose)
Docker can be used but you'd need to add the required environment variables.

### Environment Changes (Optional)
#### DNS
If you want to provide your users with a more seamless experience when working and defining provider sources', simply set `publishPublicDomains` to true and ensure the following domains resolve to the service domain.

- registry.terraform.io
