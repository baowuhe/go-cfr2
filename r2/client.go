package r2

import (
	"context"
	"fmt"
	"net/url"

	"github.com/baowuhe/go-cfr2/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// NewR2Client creates a new S3 client configured for Cloudflare R2.
func NewR2Client(cfg *config.R2Config) (*s3.Client, error) {
	// Cloudflare R2 endpoint format
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: r2Endpoint,
			// R2 does not use a region, but the AWS SDK requires one.
			// We can use a dummy region or leave it empty if the SDK allows.
			// For R2, the region is typically not relevant for endpoint resolution.
			Source: aws.EndpointSourceCustom,
		}, nil
	})

	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		awsConfig.WithEndpointResolverWithOptions(r2Resolver),
		// R2 does not use a specific region, but the SDK requires one.
		// "auto" is a common placeholder for S3-compatible storage that doesn't have regions.
		awsConfig.WithRegion("auto"), 
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	return client, nil
}

// GetR2BucketURL returns the URL for a given bucket in R2.
func GetR2BucketURL(accountID, bucketName string) string {
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", accountID, bucketName)
}

// GetR2ObjectURL returns the URL for a given object in an R2 bucket.
func GetR2ObjectURL(accountID, bucketName, objectKey string) string {
	encodedKey := url.PathEscape(objectKey)
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", accountID, bucketName, encodedKey)
}
