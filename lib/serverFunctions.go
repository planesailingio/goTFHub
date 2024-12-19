package commonlib

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v2"
)

func httpGetModules(w http.ResponseWriter, r *http.Request) {
	rootDir := filepath.Join(Prefix, "modules")
	yamlOutput, err := traverseAndBuildModuleYAML(s3fsys, rootDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error building YAML: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/yaml")
	w.Write([]byte(yamlOutput))
}

func initModuleServer() {
	r.Get(ModuleBasePath, httpGetModules)
	r.Get(ModuleBasePath+"{namespace}/{name}/{provider}/versions", httpGetModuleVersions)
	r.Get(ModuleBasePath+"{namespace}/{name}/{provider}/{version}/download", httpGetModuleDownloadURL)
}

func httpGetModuleVersions(w http.ResponseWriter, r *http.Request) {

	m := Module{
		Namespace: fmt.Sprintf("%s/%s/%s", chi.URLParam(r, "namespace"), chi.URLParam(r, "name"), chi.URLParam(r, "provider")),
		// Name:      chi.URLParam(r, "name"),
		// Provider:  chi.URLParam(r, "provider"),
	}

	// https: //registry.terraform.io/v1/modules/terraform-aws-modules/s3-bucket/aws/versions
	modPath := filepath.Join(Prefix, "modules", m.Namespace)

	modVers, err := getModuleVersions(modPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(modVers)
}

func httpGetModuleDownloadURL(w http.ResponseWriter, r *http.Request) {
	m := Module{
		Namespace: fmt.Sprintf("%s/%s/%s", chi.URLParam(r, "namespace"), chi.URLParam(r, "name"), chi.URLParam(r, "provider")),
		// Name:      chi.URLParam(r, "name"),
		// Provider:  chi.URLParam(r, "provider"),
		Version: chi.URLParam(r, "version"),
	}

	namespace := strings.ReplaceAll(m.Namespace, "/", "-")
	tfGetHeader := fmt.Sprintf("%s://%s/download/modules/%s/%s/%s-%s.tar.gz", r.Header.Get("X-Forwarded-Proto"), r.Header.Get("X-Forwarded-Host"), m.Namespace, m.Version, namespace, m.Version)
	// tfGetHeader := filepath.Join("/download", "modules", m.Namespace, m.Version, ".tar.gz")
	// tfGetHeader := filepath.Join(
	// 	"/download", "modules", m.Namespace, m.Name, m.Provider, m.Version, m.Name+".tgz",
	// )
	w.Header().Set("X-Terraform-Get", tfGetHeader)
	w.WriteHeader(http.StatusNoContent)
}

type ModuleMap map[string]map[string]map[string][]string

func traverseAndBuildModuleYAML(fsys fs.FS, rootDir string) (string, error) {
	moduleMap := make(ModuleMap)

	// Read the provider level directories
	namespaceEntries, err := fs.ReadDir(fsys, rootDir)
	if err != nil {
		return "", fmt.Errorf("error reading directory %s: %w", rootDir, err)
	}

	for _, nsEntry := range namespaceEntries {
		if !nsEntry.IsDir() {
			continue
		}
		namespacePath := filepath.Join(rootDir, nsEntry.Name())
		moduleMap[nsEntry.Name()] = make(map[string]map[string][]string)

		// Read the type level directories
		typeEntries, err := fs.ReadDir(fsys, namespacePath)
		if err != nil {
			return "", fmt.Errorf("error reading directory %s: %w", namespacePath, err)
		}

		for _, typeEntry := range typeEntries {
			if !typeEntry.IsDir() {
				continue
			}

			typeName := typeEntry.Name()
			typePath := filepath.Join(namespacePath, typeName)
			moduleMap[nsEntry.Name()][typeName] = make(map[string][]string)

			// Read the x level directories (x is another folder in the type folder)
			xEntries, err := fs.ReadDir(fsys, typePath)
			if err != nil {
				return "", fmt.Errorf("error reading directory %s: %w", typePath, err)
			}

			for _, xEntry := range xEntries {
				if !xEntry.IsDir() {
					continue
				}

				xPath := filepath.Join(typePath, xEntry.Name())
				versionEntries, err := fs.ReadDir(fsys, xPath)
				if err != nil {
					return "", fmt.Errorf("error reading directory %s: %w", xPath, err)
				}

				// Collect versions for this module
				versions := []string{}
				for _, versionEntry := range versionEntries {
					if versionEntry.IsDir() {
						versions = append(versions, versionEntry.Name())
					}
				}

				sort.Strings(versions) // Sort versions if needed (could sort in descending order if necessary)
				moduleMap[nsEntry.Name()][typeName][xEntry.Name()] = versions
			}
		}
	}

	// Convert the module map to YAML format
	yamlData, err := yaml.Marshal(moduleMap)
	if err != nil {
		return "", fmt.Errorf("error marshaling YAML: %w", err)
	}

	return string(yamlData), nil
}

//

func getProviderPlatform(input string) ProviderPlatform {
	// Efficiently split and extract platform details
	parts := strings.Split(strings.TrimSuffix(input, filepath.Ext(input)), "_")
	return ProviderPlatform{
		OS:   parts[2],
		Arch: parts[3],
	}
}

func initProviderServer() {
	r.Get(ProviderBasePath, httpGetProviders)
	r.Get(ProviderBasePath+"{namespace}/{type}/versions", httpGetProviderVersions)
	r.Get(ProviderBasePath+"{namespace}/{type}/{version}/download/{os}/{arch}", httpGetProviderDownloadPayload)
}

type ProviderDownloadResp struct {
	Protocols           []string                `json:"protocols"`
	OS                  string                  `json:"os"`
	Arch                string                  `json:"arch"`
	Filename            string                  `json:"filename"`
	DownloadUrl         string                  `json:"download_url"`
	SHASumsUrl          string                  `json:"shasums_url"`
	SHASumsSignatureUrl string                  `json:"shasums_signature_url"`
	SHASum              string                  `json:"shasum"`
	SigningKeys         ProviderMetaDataGPGKeys `json:"signing_keys"`
}

func httpGetProviderDownloadPayload(w http.ResponseWriter, r *http.Request) {
	provider := Provider{
		Namespace: chi.URLParam(r, "namespace"),
		Type:      chi.URLParam(r, "type"),
		Version:   chi.URLParam(r, "version"),
	}

	providerPlatform := ProviderPlatform{
		Arch: chi.URLParam(r, "arch"),
		OS:   chi.URLParam(r, "os"),
	}

	// Build the base URL and file name in one go
	baseUrl := filepath.Join("/download", "providers", provider.Namespace, provider.Type, provider.Version)
	fileName := fmt.Sprintf("terraform-provider-%s_%s_%s_%s.zip", provider.Type, provider.Version, providerPlatform.OS, providerPlatform.Arch)
	providerDownloadUrl := filepath.Join(baseUrl, fileName)

	metadataPath := filepath.Join(Prefix, "providers", provider.Namespace, provider.Type, provider.Version, fmt.Sprintf("meta_%s_%s.json", providerPlatform.OS, providerPlatform.Arch))
	sha256SumsUrl := filepath.Join(baseUrl, "sha256sums")
	sha256SumsSignatureUrl := filepath.Join(baseUrl, "sha256sums_signature")

	// Efficiently read metadata and compute SHA256
	var metadataPayload ProviderMetaData
	metaFile, err := s3fsys.Open(metadataPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer metaFile.Close()
	json.NewDecoder(metaFile).Decode(&metadataPayload)

	providerFile, err := s3fsys.Open(filepath.Join(Prefix, "providers", provider.Namespace, provider.Type, provider.Version, fileName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer providerFile.Close()
	fileHash, err := computeSHA256Sum(providerFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Prepare the response payload
	respPayload := ProviderDownloadResp{
		Protocols:           []string{"5.0"},
		OS:                  providerPlatform.OS,
		Arch:                providerPlatform.Arch,
		Filename:            fileName,
		DownloadUrl:         providerDownloadUrl,
		SHASumsUrl:          sha256SumsUrl,
		SHASumsSignatureUrl: sha256SumsSignatureUrl,
		SHASum:              fileHash,
		SigningKeys:         metadataPayload.SigningKeys,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respPayload)
}

func httpGetProviders(w http.ResponseWriter, r *http.Request) {
	rootDir := filepath.Join(Prefix, "providers")

	// Generate YAML output with optimized error handling
	yamlOutput, err := traverseAndBuildProviderYAML(s3fsys, rootDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error building YAML: %v", err), http.StatusInternalServerError)
		return
	}

	// Write the generated YAML to the response
	w.Header().Set("Content-Type", "text/yaml")
	w.Write([]byte(yamlOutput))
}

func getProviderVersions(providerPath string) (ProviderVersions, error) {
	var p ProviderVersions
	versionDirs, err := fs.ReadDir(s3fsys, providerPath)

	if err != nil {
		return p, err
	}

	// Loop over version directories
	for _, v := range versionDirs {
		providerVersionBasePath := filepath.Join(providerPath, v.Name())
		providerVersionFiles, err := fs.ReadDir(s3fsys, providerVersionBasePath)
		if err != nil {
			return p, err
		}

		// Collect platform info for each version
		var platforms []ProviderPlatform
		for _, l := range providerVersionFiles {
			if matched, _ := path.Match("*.zip", l.Name()); matched {
				platform := getProviderPlatform(l.Name())
				platforms = append(platforms, platform)
			}
		}

		p.Versions = append(p.Versions, ProviderVersion{
			Version:   v.Name(),
			Protocols: []string{"5.1"},
			Platforms: platforms,
		})
	}

	return p, nil
}

func httpGetProviderVersions(w http.ResponseWriter, r *http.Request) {
	provider := Provider{
		Namespace: chi.URLParam(r, "namespace"),
		Type:      chi.URLParam(r, "type"),
	}

	providerPath := filepath.Join(Prefix, "providers", provider.Namespace, provider.Type)
	providerVers, err := getProviderVersions(providerPath)
	if err != nil {
		http.Error(w, "no provider exists", http.StatusNotFound)

		return
	}
	// if len(providerVers) == 0 {

	// }

	// Respond with the provider versions in JSON format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providerVers)
}

type ProviderMap map[string]map[string][]string

func traverseAndBuildProviderYAML(fsys fs.FS, rootDir string) (string, error) {
	providerMap := make(ProviderMap)

	// Efficiently read directories and handle errors
	namespaceEntries, err := fs.ReadDir(fsys, rootDir)
	if err != nil {
		return "", fmt.Errorf("error reading directory %s: %w", rootDir, err)
	}

	// Loop through namespace entries
	for _, nsEntry := range namespaceEntries {
		if nsEntry.IsDir() {
			name := nsEntry.Name()
			providerMap[name] = make(map[string][]string)

			// Read type directories
			resourcePath := path.Join(rootDir, name)
			typeEntries, _ := fs.ReadDir(fsys, resourcePath)

			// Loop through type directories
			for _, typeEntry := range typeEntries {
				if typeEntry.IsDir() {
					typeName := typeEntry.Name()
					providerMap[name][typeName] = []string{}

					// Read version directories
					typePath := path.Join(resourcePath, typeName)
					versionEntries, _ := fs.ReadDir(fsys, typePath)
					for _, versionEntry := range versionEntries {
						if versionEntry.IsDir() {
							version := versionEntry.Name()
							providerMap[name][typeName] = append(providerMap[name][typeName], version)
						}
					}

					// Sort versions
					sortVersionsDesc(providerMap[name][typeName])
				}
			}
		}
	}

	// Convert provider map to YAML
	yamlData, err := yaml.Marshal(providerMap)
	if err != nil {
		return "", fmt.Errorf("error marshaling YAML: %w", err)
	}

	return string(yamlData), nil
}
