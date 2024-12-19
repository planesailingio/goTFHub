package commonlib

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/blang/semver"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	// "github.com/jszwec/s3fs/v2"
)

var r *chi.Mux

var (
	Backend string
	Bucket  string
	Region  string

	Prefix string
	Port   string
)

// // getEnv retrieves environment variables with a fallback to a default value.

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// computeSHA256Sum computes the SHA-256 checksum of the provided file.
func computeSHA256Sum(file io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to read file content for hashing: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// newS3Client initializes an AWS S3 client based on the backend configuration.
func newS3Client(bucket, backend string) (fs.FS, *s3.S3) {
	switch backend {
	case "minio":
		return newMinioClient(bucket)
	default:
		return newAWSClient(bucket)
	}
}

// httpGetS3File handles file downloads from S3.
func httpGetS3File(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Downloading file...")
	// "terraform-registry/artefacts/modules/terraform-aws-modules/4.2.2/terraform-aws-modules-4.2.2.tar.gz"/
	fileKey := chi.URLParam(r, "*")
	fileKeyPath := filepath.Join(Prefix, fileKey)
	fmt.Printf("Serving fileKeyPath: %s", fileKeyPath)
	file, err := s3fsys.Open(fileKeyPath)
	if err != nil {
		http.Error(w, "Error retrieving file from storage", http.StatusInternalServerError)
		fmt.Println("Error retrieving file:", err)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileKey)
	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, "Error writing file to response", http.StatusInternalServerError)
		fmt.Println("Error writing file:", err)
	}
}

// getModuleVersions fetches the available versions for a given module path.
func getModuleVersions(modPath string) (ModuleVersionsResp, error) {
	var versions []map[string]string
	versionDirs, err := fs.ReadDir(s3fsys, modPath)
	if err != nil {
		return ModuleVersionsResp{}, err
	}

	for _, v := range versionDirs {
		versions = append(versions, map[string]string{"version": v.Name()})
	}
	return ModuleVersionsResp{Modules: []ModuleVersions{{Versions: versions}}}, nil
}

// httpGetServiceDiscovery returns the service discovery response.
func httpGetServiceDiscovery(w http.ResponseWriter, r *http.Request) {
	s := ServiceDiscoveryResp{ModulesV1: ModuleBasePath, ProvidersV1: ProviderBasePath}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

// sortVersionsDesc sorts a slice of version strings in descending order using semver.
func sortVersionsDesc(versions []string) {
	semverVersions := make([]semver.Version, len(versions))
	for i, v := range versions {
		if parsedVersion, err := semver.Parse(v); err == nil {
			semverVersions[i] = parsedVersion
		} else {
			fmt.Printf("Skipping invalid version %s: %v", v, err)
		}
	}

	sort.Slice(semverVersions, func(i, j int) bool {
		return semverVersions[i].GT(semverVersions[j])
	})

	for i, v := range semverVersions {
		versions[i] = v.String()
	}
}

func Serve(backend, bucket, prefix, region string) {
	Backend = backend
	Bucket = bucket
	Prefix = prefix
	Region = region
	Port = "3000"

	s3fsys, s3Client = newS3Client(bucket, backend)
	providerTestPath := filepath.Join(Prefix, "providers/index.md")
	moduleTestPath := filepath.Join(Prefix, "modules/index.md")
	if _, err := fs.Stat(s3fsys, providerTestPath); err != nil {

		fmt.Printf("WARN: Connecting to storage backend. Testing provider path s3://%s/%s: %v\n", Bucket, providerTestPath, err)
		fmt.Printf("WARN: Creating new base at s3://%s/%s\n", Bucket, Prefix)
		writeFile(s3Client, Bucket, providerTestPath, "Provider Home")
		writeFile(s3Client, Bucket, moduleTestPath, "Provider Home")
	}
	fmt.Printf("Testing S3 Storage: s3://%s/%s...success!\n", bucket, providerTestPath)
	fmt.Printf("Initialising server on 0.0.0.0:%s\n", Port)
	r = chi.NewRouter()
	r.Use(middleware.RealIP, middleware.RequestID, middleware.Recoverer, middleware.Logger, middleware.GetHead)
	r.Use(middleware.Heartbeat("/is_alive"))

	r.Get("/", httpGetServiceDiscovery)
	r.Get("/.well-known/terraform.json", httpGetServiceDiscovery)
	r.Get("/download/*", httpGetS3File)

	initModuleServer()
	initProviderServer()

	http.ListenAndServe(":"+Port, r)
}
