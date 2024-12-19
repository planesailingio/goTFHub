package commonlib

// ExampleProvider represents a sample provider configuration for YAML
type ArtefactList struct {
	Providers []Provider
	Modules   []Module
}

const TerraformRegistryAPIProviderEndpoint = "https://registry.terraform.io/v1/providers"
const TerraformRegistryAPIModuleEndpoint = "https://registry.terraform.io/v1/modules"

var osArchTypes = []map[string]string{
	{"os": "linux", "arch": "amd64"},
	{"os": "linux", "arch": "arm64"},
	{"os": "darwin", "arch": "amd64"},
	{"os": "darwin", "arch": "arm64"},
	{"os": "windows", "arch": "amd64"},
	{"os": "windows", "arch": "arm64"},
}

const (
	ModuleBasePath   = "/v1/modules/"
	ProviderBasePath = "/v1/providers/"
)

type VersionInfo struct {
	Filename            string   `json:"filename"`
	DownloadURL         string   `json:"download_url"`
	ShasumsURL          string   `json:"shasums_url"`
	ShasumsSignatureURL string   `json:"shasums_signature_url"`
	Shasum              string   `json:"shasum"`
	Protocols           []string `json:"protocols"`
	SigningKeys         any      `json:"signing_keys"`
	OS                  string   `json:"os"`
	Arch                string   `json:"arch"`
}

// ServiceDiscoveryResp defines the service discovery response structure.
type ServiceDiscoveryResp struct {
	ModulesV1   string `json:"modules.v1"`
	ProvidersV1 string `json:"providers.v1"`
}

// ModuleVersions contains module versions.
type ModuleVersions struct {
	Versions []map[string]string `json:"versions"`
}

// ModuleVersionsResp defines the response structure for module versions.
type ModuleVersionsResp struct {
	Modules []ModuleVersions `json:"modules"`
}

// Module represents a terraform module.
type Module struct {
	Namespace string
	// Name      string
	// Provider  string
	Version string
}

//	type Provider struct {
//		Namespace string `yaml:"source"`
//		Version   string `yaml:"version,omitempty"`
//	}
type Provider struct {
	Namespace string `yaml:"namespace"`
	Type      string `yaml:"type,omitempty"`
	Version   string `yaml:"version,omitempty"`
}

type GPGPublicKey struct {
	KeyId      string `json:"key_id"`
	AsciiArmor string `json:"ascii_armor"`
	Source     string `json:"source"`
	SourceUrl  string `json:"source_url"`
}

type ProviderMetaDataGPGKeys struct {
	GPGPublicKeys []GPGPublicKey `json:"gpg_public_keys"`
}

type ProviderMetaData struct {
	Shasum      string                  `json:"shasum"`
	SigningKeys ProviderMetaDataGPGKeys `json:"signing_keys"`
}

type ProviderVersions struct {
	Versions []ProviderVersion `json:"versions"`
}

type ProviderVersion struct {
	Version   string             `json:"version"`
	Protocols []string           `json:"protocols"`
	Platforms []ProviderPlatform `json:"platforms"`
}

type ProviderPlatform struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}
