package r2

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ListObjects lists all objects in the specified R2 bucket.
func ListObjects(ctx context.Context, client *s3.Client, bucketName string) ([]types.Object, error) {
	var allObjects []types.Object
	input := &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	}

	paginator := s3.NewListObjectsV2Paginator(client, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}
		allObjects = append(allObjects, output.Contents...)
	}

	return allObjects, nil
}

// DeleteObject deletes an object from the specified R2 bucket.
func DeleteObject(ctx context.Context, client *s3.Client, bucketName, objectKey string) error {
	input := &s3.DeleteObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	}

	_, err := client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object '%s' from bucket '%s': %w", objectKey, bucketName, err)
	}

	return nil
}

// RenameObject renames an object in the specified R2 bucket by copying it to a new key and deleting the original.
func RenameObject(ctx context.Context, client *s3.Client, bucketName, oldObjectKey, newObjectKey string) error {
	// First, copy the object to the new key
	copyInput := &s3.CopyObjectInput{
		Bucket:     &bucketName,
		CopySource: aws.String(bucketName + "/" + oldObjectKey),
		Key:        &newObjectKey,
	}

	_, err := client.CopyObject(ctx, copyInput)
	if err != nil {
		return fmt.Errorf("failed to copy object from '%s' to '%s' in bucket '%s': %w", oldObjectKey, newObjectKey, bucketName, err)
	}

	// Then, delete the original object
	err = DeleteObject(ctx, client, bucketName, oldObjectKey)
	if err != nil {
		// If deletion fails, return an error but the rename has already happened at the copy stage
		return fmt.Errorf("copy successful but failed to delete original object '%s' from bucket '%s': %w", oldObjectKey, bucketName, err)
	}

	return nil
}

// progressWriter is a custom io.Writer that reports progress for downloads.
type progressWriter struct {
	io.Writer
	total       int64
	transferred int64
	mu          sync.Mutex
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.mu.Lock()
	pw.transferred += int64(n)
	pw.mu.Unlock()

	// Print progress on a single line
	percentage := float64(pw.transferred) / float64(pw.total) * 100
	fmt.Fprintf(os.Stdout, "\r%d / %d (%.2f%%)", pw.transferred, pw.total, percentage)
	os.Stdout.Sync() // Ensure immediate flush
	return n, nil
}

// progressReader is a custom io.Reader that reports progress for uploads.
type progressReader struct {
	io.Reader
	total       int64
	transferred int64
	mu          sync.Mutex
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	if err != nil {
		return n, err
	}

	pr.mu.Lock()
	pr.transferred += int64(n)
	pr.mu.Unlock()

	// Print progress on a single line
	percentage := float64(pr.transferred) / float64(pr.total) * 10
	fmt.Fprintf(os.Stdout, "\r%d / %d (%.2f%%)", pr.transferred, pr.total, percentage)
	os.Stdout.Sync() // Ensure immediate flush
	return n, nil
}

// DownloadObject downloads an object from the specified R2 bucket to a local file.
func DownloadObject(ctx context.Context, client *s3.Client, bucketName, objectKey, localFilePath string) error {
	input := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	}

	resp, err := client.GetObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to get object '%s' from bucket '%s': %w", objectKey, bucketName, err)
	}
	defer resp.Body.Close()

	file, err := os.Create(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to create local file '%s': %w", localFilePath, err)
	}
	defer file.Close()

	// Get total size for progress tracking
	var totalSize int64
	if resp.ContentLength != nil {
		totalSize = *resp.ContentLength
	} else {
		fmt.Println("Warning: ContentLength not available, download progress percentage will not be shown.")
	}

	pw := &progressWriter{
		Writer: file,
		total:  totalSize,
	}

	_, err = io.Copy(pw, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write object content to file '%s': %w", localFilePath, err)
	}
	fmt.Println() // Newline after download completes

	return nil
}

// UploadObject uploads a local file to the specified R2 bucket.
func UploadObject(ctx context.Context, client *s3.Client, bucketName, objectKey, localFilePath string) error {
	file, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to open local file '%s': %w", localFilePath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info for '%s': %w", localFilePath, err)
	}
	fileSize := fileInfo.Size()

	pr := &progressReader{
		Reader: file,
		total:  fileSize,
	}

	uploader := manager.NewUploader(client)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
		Body:   pr, // Use progressReader as the Body
	})
	if err != nil {
		return fmt.Errorf("failed to upload object '%s' to bucket '%s': %w", objectKey, bucketName, err)
	}
	fmt.Println() // Newline after upload completes

	return nil
}

// GeneratePresignedURL generates a presigned URL for an object in the specified R2 bucket with a default expiration of 24 hours.
func GeneratePresignedURL(ctx context.Context, client *s3.Client, bucketName, objectKey string) (string, error) {
	return GeneratePresignedURLWithExpiry(ctx, client, bucketName, objectKey, 24*time.Hour)
}

// GeneratePresignedURLWithExpiry generates a presigned URL for an object in the specified R2 bucket with a custom expiration time.
func GeneratePresignedURLWithExpiry(ctx context.Context, client *s3.Client, bucketName, objectKey string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(client) // Correct usage of NewPresignClient

	input := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	}

	result, err := presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL for object '%s' in bucket '%s': %w", objectKey, bucketName, err)
	}

	return result.URL, nil
}
