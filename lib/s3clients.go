package commonlib

import (
	"crypto/tls"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"

	// "github.com/aws/aws-sdk-go/service/s3"
	"github.com/packrat386/s3fs"
)

var s3fsys fs.FS
var s3Client *s3.S3

// newAWSClient initializes an AWS S3 client.
func newAWSClient(bucket string) (fs.FS, *s3.S3) {
	sessionOptions := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}
	sess, _ := session.NewSessionWithOptions(sessionOptions)
	if *sess.Config.Region == "" {
		*sess.Config.Region = Region
		fmt.Printf("WARN: AWS Region not set. Using default: %s\n", Region)
	}
	client := s3.New(sess)
	return s3fs.NewS3FS(client, bucket), client
}

var minio_endpoint, minio_access_key_id, minio_secret_access_key string
var s3_tls_enable, s3_tls_skip_verify bool

// newMinioClient initializes a MinIO-compatible S3 client.
func newMinioClient(bucket string) (fs.FS, *s3.S3) {
	minio_endpoint = getEnv("MINIO_ENDPOINT", "http://localhost:9001")
	minio_access_key_id = getEnv("MINIO_ACCESS_KEY_ID", "minioadmin")
	minio_secret_access_key = getEnv("MINIO_SECRET_ACCESS_KEY", "minioadmin")
	Region = getEnv("MINIO_REGION", "us-east-1")
	s3_tls_enable, _ = strconv.ParseBool(getEnv("MINIO_TLS_ENABLE", "true"))
	s3_tls_enable = !s3_tls_enable
	s3_tls_skip_verify, _ = strconv.ParseBool(getEnv("MINIO_TLS_SKIP_VERIFY", "false"))

	customTransport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: s3_tls_skip_verify}}
	customHTTPClient := &http.Client{Transport: customTransport}
	sess, _ := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(minio_access_key_id, minio_secret_access_key, ""),
		Endpoint:         &minio_endpoint,
		Region:           &Region,
		DisableSSL:       &s3_tls_enable,
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       customHTTPClient,
	})
	client := s3.New(sess)
	return s3fs.NewS3FS(client, bucket), client
}

func writeFile(client *s3.S3, bucket, key, body string) {
	_, err := client.PutObject(&s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(strings.NewReader(body)),
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		panic(err)
	}
}

// UploadDirectory uploads a local directory recursively to an S3 bucket
func UploadDirectory(backend, bucket, localPath, prefix string) error {
	Prefix = prefix
	Bucket = bucket
	Backend = backend
	// var s3Client *s3.S3
	_, s3Client := newS3Client(bucket, backend)
	fmt.Printf("Scanning path: %s\n", localPath)

	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		fmt.Println(path)
		if err != nil {
			return fmt.Errorf("error accessing path %s: %v", path, err)
		}

		// Skip directories (upload files only)
		if info.IsDir() {
			return nil
		}

		// Derive S3 key by trimming local path and adding s3Prefix
		relativePath := strings.TrimPrefix(path, localPath)
		s3Key := strings.TrimPrefix(filepath.Join(Prefix, relativePath), string(filepath.Separator))

		// Open the file
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", path, err)
		}
		defer file.Close()

		// Upload the file
		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(s3Key),
			Body:   file,
		})
		if err != nil {
			return fmt.Errorf("failed to upload %s to s3://%s/%s: %v", path, bucket, s3Key, err)
		}

		fmt.Printf("Uploaded %s to s3://%s/%s\n", path, bucket, s3Key)
		return nil
	})

	return err
}
