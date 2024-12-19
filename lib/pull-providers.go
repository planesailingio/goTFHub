package commonlib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

func LoadProviders(file string) ([]Provider, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var providers struct {
		Providers []Provider `yaml:"providers"`
	}
	if err := yaml.Unmarshal(data, &providers); err != nil {
		return nil, err
	}
	return providers.Providers, nil
}

func ProcessProvider(provider Provider, providerCount int, outputDir string) error {

	var selectedVersions []string
	if provider.Version != "" {
		selectedVersions = []string{provider.Version}
	} else {
		versions, err := FetchProviderVersions(provider.Namespace)

		if err != nil {
			return err
		}

		selectedVersions = versions
		if len(versions) > providerCount {
			selectedVersions = versions[:providerCount]
		}
	}
	for _, version := range selectedVersions {
		fmt.Printf("Fetching Provider: %s:%s\n", provider.Namespace, version)
		if err := DownloadVersion(provider.Namespace, version, outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to download version %s: %v\n", version, err)
		}
	}
	return nil
}

func FetchProviderVersions(source string) ([]string, error) {
	url := fmt.Sprintf("%s/%s/versions", TerraformRegistryAPIProviderEndpoint, source)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	var result struct {
		Versions []struct {
			Version string `json:"version"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var versions []string
	for _, v := range result.Versions {
		versions = append(versions, v.Version)
	}
	sort.Slice(versions, func(i, j int) bool {
		v1, err1 := semver.NewVersion(versions[i])
		v2, err2 := semver.NewVersion(versions[j])
		if err1 != nil || err2 != nil {
			// Handle parse errors (optional, depending on your requirements)
			return false
		}
		// Compare versions in descending order
		return v1.GreaterThan(v2)
	})
	// spew.Dump(versions)

	// sort.Strings(versions)
	// slices.Reverse(versions)
	// sort.strings(versions, func(i, j int) bool {
	// 	return versions[i] > versions[j]
	// })
	// fmt.Println(versions)
	return versions, nil
}

func FilterVersions(versions []string, targetVersion string, count int) []string {
	// targetVersion = "3.9.1"
	if targetVersion != "" {
		for _, v := range versions {
			if v == targetVersion {
				return []string{v}
			}
		}
		return []string{} // No match found
	} else {
		fmt.Println("Return latest")
		// spew.Dump(versions)

	}

	if len(versions) > count {
		return versions[:count]
	}
	return versions
}

func DownloadVersion(source, version, outputDir string) error {
	var wg sync.WaitGroup
	for _, osArch := range osArchTypes {
		wg.Add(1)
		go func(osArch map[string]string) {
			defer wg.Done()
			if err := DownloadOSArch(source, version, osArch["os"], osArch["arch"], outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Failed for %s/%s: %v\n", osArch["os"], osArch["arch"], err)
			}
		}(osArch)
	}
	wg.Wait()
	return nil
}

func DownloadOSArch(source, version, osName, arch, outputDir string) error {
	url := fmt.Sprintf("%s/%s/%s/download/%s/%s", TerraformRegistryAPIProviderEndpoint, source, version, osName, arch)
	// fmt.Printf("Downloading %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// fmt.Printf("RESPONSE CODE: %d\n", resp.StatusCode)
	if resp.StatusCode == http.StatusNotFound {
		// fmt.Printf("%s/%s not found, skipping.\n", osName, arch)
		return nil
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed with status: %d", resp.StatusCode)
	}

	outputPath := filepath.Join(outputDir, source, version, fmt.Sprintf("meta_%s_%s.json", osName, arch))

	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return err
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(resp.Body)
	file.Write(data)

	if err == nil {
		// fmt.Printf("Downloaded meta %s/%s for %s\n", osName, arch, version)
		var versionInfo VersionInfo
		err = json.Unmarshal(data, &versionInfo)
		binaryPath := filepath.Join(outputDir, source, version, versionInfo.Filename)
		errDownloadFile := downloadFile(versionInfo.DownloadURL, binaryPath)
		if errDownloadFile != nil {
			fmt.Printf("ERROR: %s", errDownloadFile.Error())
		}

		errDownloadFile = downloadFile(versionInfo.ShasumsURL, filepath.Join(outputDir, source, version, "sha256sums"))
		if errDownloadFile != nil {
			fmt.Printf("ERROR: %s", errDownloadFile.Error())
		}

		errDownloadFile = downloadFile(versionInfo.ShasumsSignatureURL, filepath.Join(outputDir, source, version, "sha256sums_signature"))
		if errDownloadFile != nil {
			fmt.Printf("ERROR: %s", errDownloadFile.Error())
		}

	}

	return err
}
func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	// fmt.Printf("Downloaded: %s\n", destPath)
	if err != nil {
		return fmt.Errorf("failed to save downloaded file: %w", err)
	}

	return nil
}
