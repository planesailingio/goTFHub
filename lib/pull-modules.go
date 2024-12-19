package commonlib

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
)

// func LoadProviders(file string) ([]Provider, error) {
// 	data, err := os.ReadFile(file)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var modules struct {
// 		Providers []Provider `yaml:"modules"`
// 	}
// 	if err := yaml.Unmarshal(data, &modules); err != nil {
// 		return nil, err
// 	}
// 	return modules.Providers, nil
// }

func LoadModules(file string) ([]Module, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var modules struct {
		Modules []Module `yaml:"modules"`
	}
	if err := yaml.Unmarshal(data, &modules); err != nil {
		return nil, err
	}
	return modules.Modules, nil
}

func ProcessModule(module Module, moduleCount int, outputDir string) error {

	var selectedVersions []string
	if module.Version != "" {
		selectedVersions = []string{module.Version}
	} else {
		versions, err := FetchModuleVersions(module.Namespace)

		if err != nil {
			return err
		}

		selectedVersions = versions
		if len(versions) > moduleCount {
			selectedVersions = versions[:moduleCount]
		}
	}
	// selectedVersions := FilterVersions(versions, module.Version, moduleCount)
	for _, version := range selectedVersions {
		fmt.Printf("Fetching Module: %s:%s\n", module.Namespace, version)
		if err := DownloadModuleVersion(module.Namespace, version, outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to download version %s: %v\n", version, err)
		}
	}
	return nil
}

func FetchModuleVersions(source string) ([]string, error) {
	// https: //registry.terraform.io/v1/modules/terraform-aws-modules/s3-bucket/aws/versions
	url := fmt.Sprintf("%s/%s/versions", TerraformRegistryAPIModuleEndpoint, source)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}
	// spew.Dump(resp.StatusCode)

	var result struct {
		Modules []struct {
			Source   string `json:"source"`
			Versions []struct {
				Version string `json:"version"`
				Root    struct {
					Providers []struct {
						Name      string `json:"name"`
						Namespace string `json:"namespace"`
						Source    string `json:"source"`
						Version   string `json:"version"`
					} `json:"providers"`
					// Dependencies []string `json:"dependencies"`
				} `json:"root"`
			} `json:"versions"`
		} `json:"modules"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	// fmt.Printf("Version Count: %d\n", len(result.Modules[0].Versions))
	var versions []string
	for _, v := range result.Modules[0].Versions {
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
	return versions, nil
}

// func FilterVersions(versions []string, targetVersion string, count int) []string {
// 	// targetVersion = "3.9.1"
// 	if targetVersion != "" {
// 		for _, v := range versions {
// 			if v == targetVersion {
// 				return []string{v}
// 			}
// 		}
// 		return []string{} // No match found
// 	} else {
// 		fmt.Println("Return latest")
// 		// spew.Dump(versions)

// 	}

// 	if len(versions) > count {
// 		return versions[:count]
// 	}
// 	return versions
// }

func DownloadModuleVersion(source, version, outputDir string) error {
	url := fmt.Sprintf("%s/%s/%s/download", TerraformRegistryAPIModuleEndpoint, source, version)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// fmt.Printf("RESPONSE CODE: %d\n", resp.StatusCode)
	if resp.StatusCode == http.StatusNotFound {
		// fmt.Printf("%s/%s not found, skipping.\n", osName, arch)
		return nil
	} else if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed with status: %d", resp.StatusCode)
	}

	repoURL := resp.Header.Get("x-terraform-get")
	basePath := filepath.Join(outputDir, source, version)
	outputFile := filepath.Join(basePath, strings.ReplaceAll(source, "/", "-")+"-"+version+".tar.gz")
	if err := os.MkdirAll(basePath, 0755); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	switch repoURL[:5] {
	case "git::":
		if err := CloneAndCompress(repoURL, outputFile); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	default:
		fmt.Printf("ERROR: Module source not support: %s\n", repoURL)
	}

	// var wg sync.WaitGroup
	// for _, osArch := range osArchTypes {
	// 	wg.Add(1)
	// 	go func(osArch map[string]string) {
	// 		defer wg.Done()
	// 		if err := DownloadOSArch(source, version, osArch["os"], osArch["arch"], outputDir); err != nil {
	// 			fmt.Fprintf(os.Stderr, "Failed for %s/%s: %v\n", osArch["os"], osArch["arch"], err)
	// 		}
	// 	}(osArch)
	// }
	// wg.Wait()
	return nil
}

// func ProcessModuleGit(url string) string {

// }

// func DownloadModuleVersion(source, version, outputDir string) error {
// 	url := fmt.Sprintf("%s/%s/%s/download", TerraformRegistryAPIModuleEndpoint, source, version)
// 	fmt.Printf("Downloading %s\n", url)
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()
// 	// fmt.Printf("RESPONSE CODE: %d\n", resp.StatusCode)
// 	if resp.StatusCode == http.StatusNotFound {
// 		// fmt.Printf("%s/%s not found, skipping.\n", osName, arch)
// 		return nil
// 	} else if resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("failed with status: %d", resp.StatusCode)
// 	}

// 	outputPath := filepath.Join(outputDir, source, version, fmt.Sprintf("meta_%s_%s.json", osName, arch))

// 	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
// 		return err
// 	}
// 	file, err := os.Create(outputPath)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	data, err := io.ReadAll(resp.Body)
// 	file.Write(data)

// 	if err == nil {
// 		// fmt.Printf("Downloaded meta %s/%s for %s\n", osName, arch, version)
// 		var versionInfo VersionInfo
// 		err = json.Unmarshal(data, &versionInfo)
// 		binaryPath := filepath.Join(outputDir, source, version, versionInfo.Filename)
// 		errDownloadFile := downloadFile(versionInfo.DownloadURL, binaryPath)
// 		if errDownloadFile != nil {
// 			fmt.Printf("ERROR: %s", errDownloadFile.Error())
// 		}

// 		errDownloadFile = downloadFile(versionInfo.ShasumsURL, filepath.Join(outputDir, source, version, "sha256sums"))
// 		if errDownloadFile != nil {
// 			fmt.Printf("ERROR: %s", errDownloadFile.Error())
// 		}

// 		errDownloadFile = downloadFile(versionInfo.ShasumsSignatureURL, filepath.Join(outputDir, source, version, "sha256sums_signature"))
// 		if errDownloadFile != nil {
// 			fmt.Printf("ERROR: %s", errDownloadFile.Error())
// 		}

// 	}

// 	return err
// }

// func downloadFile(url, destPath string) error {
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return fmt.Errorf("failed to download file: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	file, err := os.Create(destPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to create file: %w", err)
// 	}
// 	defer file.Close()

// 	_, err = io.Copy(file, resp.Body)
// 	fmt.Printf("Downloaded: %s\n", destPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to save downloaded file: %w", err)
// 	}

// 	return nil
// }

func CloneAndCompress(repoURL, outputFile string) error {
	// Parse the URL to extract the actual Git URL and reference (if any)
	parts := strings.Split(repoURL, "::")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository URL format")
	}
	gitURL := parts[1]
	ref := ""
	if idx := strings.Index(gitURL, "?ref="); idx != -1 {
		ref = gitURL[idx+5:]
		gitURL = gitURL[:idx]
	}

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "git-clone-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir) // Ensure cleanup

	// Clone the repository
	cloneCmd := exec.Command("git", "clone", gitURL, tmpDir)
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Checkout the specific ref (if provided)
	if ref != "" {
		checkoutCmd := exec.Command("git", "checkout", ref)
		checkoutCmd.Dir = tmpDir
		if err := checkoutCmd.Run(); err != nil {
			return fmt.Errorf("failed to checkout ref %s: %w", ref, err)
		}
	}

	// Create the tar.gz file
	err = createTarGz(outputFile, tmpDir)
	if err != nil {
		return fmt.Errorf("failed to create tar.gz file: %w", err)
	}

	// fmt.Printf("Repository successfully cloned, compressed, and saved to %s\n", outputFile)
	return nil
}

// createTarGz compresses the source directory into a tar.gz file
func createTarGz(outputFile, sourceDir string) error {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk through the source directory and add files to the tar.gz archive
	err = filepath.Walk(sourceDir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create the header
		header, err := tar.FileInfoHeader(info, file)
		if err != nil {
			return err
		}

		// Update the name to maintain relative paths
		header.Name, err = filepath.Rel(sourceDir, file)
		if err != nil {
			return err
		}

		// Write the header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If the file is not a directory, write its content
		if !info.IsDir() {
			fileContent, err := os.Open(file)
			if err != nil {
				return err
			}
			defer fileContent.Close()

			if _, err := io.Copy(tarWriter, fileContent); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error during compression: %w", err)
	}

	return nil
}
