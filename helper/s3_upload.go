package helper

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Uploader uploads files to an S3 bucket and returns public URLs.
type S3Uploader struct {
	client        *s3.Client
	bucket        string
	region        string
	prefix        string
	publicBaseURL string
}

// NewS3UploaderFromEnv builds an uploader from environment variables.
//
// Required: S3_BUCKET, AWS_REGION (or S3_REGION)
// Optional: S3_PREFIX (default "uploads"), S3_PUBLIC_BASE_URL (CloudFront/custom domain),
// AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY (otherwise default AWS credential chain).
func NewS3UploaderFromEnv() (*S3Uploader, error) {
	bucket := strings.TrimSpace(os.Getenv("S3_BUCKET"))
	if bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET is required")
	}

	region := strings.TrimSpace(os.Getenv("S3_REGION"))
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_REGION"))
	}
	if region == "" {
		return nil, fmt.Errorf("AWS_REGION or S3_REGION is required")
	}

	prefix := strings.Trim(strings.TrimSpace(os.Getenv("S3_PREFIX")), "/")
	if prefix == "" {
		prefix = "uploads"
	}

	publicBase := strings.TrimRight(strings.TrimSpace(os.Getenv("S3_PUBLIC_BASE_URL")), "/")

	cfgOpts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	accessKey := strings.TrimSpace(os.Getenv("AWS_ACCESS_KEY_ID"))
	secretKey := strings.TrimSpace(os.Getenv("AWS_SECRET_ACCESS_KEY"))
	if accessKey != "" && secretKey != "" {
		cfgOpts = append(cfgOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), cfgOpts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	return &S3Uploader{
		client:        s3.NewFromConfig(cfg),
		bucket:        bucket,
		region:        region,
		prefix:        prefix,
		publicBaseURL: publicBase,
	}, nil
}

// Upload stores a file under folder/filename and returns its public URL.
func (u *S3Uploader) Upload(ctx context.Context, body io.Reader, contentType, folder, filename string) (string, error) {
	key := u.objectKey(folder, filename)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	}

	if _, err := u.client.PutObject(ctx, input); err != nil {
		return "", fmt.Errorf("s3 put object: %w", err)
	}

	return u.publicURL(key), nil
}

func (u *S3Uploader) objectKey(folder, filename string) string {
	parts := make([]string, 0, 3)
	if u.prefix != "" {
		parts = append(parts, u.prefix)
	}
	if folder != "" {
		parts = append(parts, strings.Trim(folder, "/"))
	}
	parts = append(parts, filename)
	return strings.Join(parts, "/")
}

func (u *S3Uploader) publicURL(key string) string {
	if u.publicBaseURL != "" {
		return u.publicBaseURL + "/" + key
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", u.bucket, u.region, key)
}

// ContentTypeForFilename guesses a MIME type from the file extension.
func ContentTypeForFilename(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "application/octet-stream"
	}
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	return "application/octet-stream"
}
