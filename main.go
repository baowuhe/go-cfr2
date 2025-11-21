package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/baowuhe/go-cfr2/config"
	"github.com/baowuhe/go-cfr2/r2"
	"github.com/baowuhe/go-cfr2/utils"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	cfg, err := config.LoadConfig()
	if err != nil {
	utils.ExitWithError(fmt.Sprintf("Configuration error: %v", err))
	}

	client, err := r2.NewR2Client(cfg)
	if err != nil {
		utils.ExitWithError(fmt.Sprintf("Failed to create R2 client: %v", err))
	}

	switch command {
	case "list":
		handleListCommand(context.Background(), client, cfg)
	case "download":
		handleDownloadCommand(context.Background(), client, cfg)
	case "upload":
		handleUploadCommand(context.Background(), client, cfg)
	case "delete":
		handleDeleteCommand(context.Background(), client, cfg)
	case "rename":
		handleRenameCommand(context.Background(), client, cfg)
	case "presign":
		handlePresignCommand(context.Background(), client, cfg)
	default:
		printUsage()
		os.Exit(1)
	}
}

func handleListCommand(ctx context.Context, client *s3.Client, cfg *config.R2Config) {
	listFlags := flag.NewFlagSet("list", flag.ExitOnError)
	bucketName := listFlags.String("b", cfg.DefaultBucket, "Specify the R2 bucket name (optional)")
	listFlags.StringVar(bucketName, "bucket", cfg.DefaultBucket, "Specify the R2 bucket name (optional)")
	listFlags.Parse(os.Args[2:])

	if *bucketName == "" {
		utils.ExitWithError("Bucket name not specified. Use -b or --bucket flag, or set DefaultBucket in config.")
	}

	objects, err := r2.ListObjects(ctx, client, *bucketName)
	if err != nil {
	utils.ExitWithError(fmt.Sprintf("Failed to list objects in bucket '%s': %v", *bucketName, err))
	}

	if len(objects) == 0 {
		fmt.Println("No objects found in the bucket.")
		return
	}

	for _, obj := range objects {
	sizeStr := "N/A"
		if obj.Size != nil {
			sizeStr = strconv.FormatInt(*obj.Size, 10)
		}
		fmt.Printf("%s | %s\n", *obj.Key, sizeStr)
	}
}

func handleDownloadCommand(ctx context.Context, client *s3.Client, cfg *config.R2Config) {
	downloadFlags := flag.NewFlagSet("download", flag.ExitOnError)
	bucketName := downloadFlags.String("b", cfg.DefaultBucket, "Specify the R2 bucket name (optional)")
	downloadFlags.StringVar(bucketName, "bucket", cfg.DefaultBucket, "Specify the R2 bucket name (optional)")
	objectKey := downloadFlags.String("k", "", "Specify the object key to download (required)")
	downloadFlags.StringVar(objectKey, "key", "", "Specify the object key to download (required)")
	outputPath := downloadFlags.String("o", "", "Specify the output file path or directory (optional)")
	downloadFlags.StringVar(outputPath, "output", "", "Specify the output file path or directory (optional)")
	downloadFlags.Parse(os.Args[2:])

	if *bucketName == "" {
		utils.ExitWithError("Bucket name not specified. Use -b or --bucket flag, or set DefaultBucket in config.")
	}
	if *objectKey == "" {
		utils.ExitWithError("Object key not specified. Use -k or --key flag.")
	}

	finalOutputPath := *outputPath
	if finalOutputPath == "" {
		// Default to current directory, replace '/' in key with '_'
		fileName := strings.ReplaceAll(*objectKey, "/", "_")
	finalOutputPath = filepath.Join(".", fileName)
	} else {
		// If output is a directory, append the filename from objectKey
	if stat, err := os.Stat(finalOutputPath); err == nil && stat.IsDir() {
			fileName := filepath.Base(*objectKey)
			finalOutputPath = filepath.Join(finalOutputPath, fileName)
		}
	}

	fmt.Printf("Downloading '%s' from bucket '%s' to '%s'...\n", *objectKey, *bucketName, finalOutputPath)
	err := r2.DownloadObject(ctx, client, *bucketName, *objectKey, finalOutputPath)
	if err != nil {
	utils.ExitWithError(fmt.Sprintf("Failed to download object '%s': %v", *objectKey, err))
	}
	fmt.Printf("Successfully downloaded '%s' to '%s'.\n", *objectKey, finalOutputPath)
}

func handleUploadCommand(ctx context.Context, client *s3.Client, cfg *config.R2Config) {
	uploadFlags := flag.NewFlagSet("upload", flag.ExitOnError)
	bucketName := uploadFlags.String("b", cfg.DefaultBucket, "Specify the R2 bucket name (optional)")
	uploadFlags.StringVar(bucketName, "bucket", cfg.DefaultBucket, "Specify the R2 bucket name (optional)")
	filePath := uploadFlags.String("f", "", "Specify the local file to upload (required)")
	uploadFlags.StringVar(filePath, "file", "", "Specify the local file to upload (required)")
	objectKey := uploadFlags.String("k", "", "Specify the object key for the uploaded file (required)")
	uploadFlags.StringVar(objectKey, "key", "", "Specify the object key for the uploaded file (required)")
	uploadFlags.Parse(os.Args[2:])

	if *bucketName == "" {
		utils.ExitWithError("Bucket name not specified. Use -b or --bucket flag, or set DefaultBucket in config.")
	}
	if *filePath == "" {
	utils.ExitWithError("File path not specified. Use -f or --file flag.")
	}
	if *objectKey == "" {
		utils.ExitWithError("Object key not specified. Use -k or --key flag.")
	}

	fmt.Printf("Uploading '%s' to bucket '%s' as '%s'...\n", *filePath, *bucketName, *objectKey)
	err := r2.UploadObject(ctx, client, *bucketName, *objectKey, *filePath)
	if err != nil {
		utils.ExitWithError(fmt.Sprintf("Failed to upload file '%s': %v", *filePath, err))
	}
	fmt.Printf("Successfully uploaded '%s' to '%s'.\n", *filePath, *objectKey)
}

func handleDeleteCommand(ctx context.Context, client *s3.Client, cfg *config.R2Config) {
	deleteFlags := flag.NewFlagSet("delete", flag.ExitOnError)
	bucketName := deleteFlags.String("b", cfg.DefaultBucket, "Specify the R2 bucket name (optional)")
	deleteFlags.StringVar(bucketName, "bucket", cfg.DefaultBucket, "Specify the R2 bucket name (optional)")
	objectKey := deleteFlags.String("k", "", "Specify the object key to delete (required)")
	deleteFlags.StringVar(objectKey, "key", "", "Specify the object key to delete (required)")
	deleteFlags.Parse(os.Args[2:])

	if *bucketName == "" {
		utils.ExitWithError("Bucket name not specified. Use -b or --bucket flag, or set DefaultBucket in config.")
	}
	if *objectKey == "" {
		utils.ExitWithError("Object key not specified. Use -k or --key flag.")
	}

	fmt.Printf("Deleting '%s' from bucket '%s'...\n", *objectKey, *bucketName)
	err := r2.DeleteObject(ctx, client, *bucketName, *objectKey)
	if err != nil {
	utils.ExitWithError(fmt.Sprintf("Failed to delete object '%s': %v", *objectKey, err))
	}
	fmt.Printf("Successfully deleted '%s' from '%s'.\n", *objectKey, *bucketName)
}

func handleRenameCommand(ctx context.Context, client *s3.Client, cfg *config.R2Config) {
	renameFlags := flag.NewFlagSet("rename", flag.ExitOnError)
	bucketName := renameFlags.String("b", cfg.DefaultBucket, "Specify the R2 bucket name (optional)")
	renameFlags.StringVar(bucketName, "bucket", cfg.DefaultBucket, "Specify the R2 bucket name (optional)")
	oldObjectKey := renameFlags.String("o", "", "Specify the old object key to rename (required)")
	renameFlags.StringVar(oldObjectKey, "old-key", "", "Specify the old object key to rename (required)")
	newObjectKey := renameFlags.String("n", "", "Specify the new object key (required)")
	renameFlags.StringVar(newObjectKey, "new-key", "", "Specify the new object key (required)")
	renameFlags.Parse(os.Args[2:])

	if *bucketName == "" {
		utils.ExitWithError("Bucket name not specified. Use -b or --bucket flag, or set DefaultBucket in config.")
	}
	if *oldObjectKey == "" {
		utils.ExitWithError("Old object key not specified. Use -old or --old-key flag.")
	}
	if *newObjectKey == "" {
		utils.ExitWithError("New object key not specified. Use -new or --new-key flag.")
	}

	fmt.Printf("Renaming '%s' to '%s' in bucket '%s'...\n", *oldObjectKey, *newObjectKey, *bucketName)
	err := r2.RenameObject(ctx, client, *bucketName, *oldObjectKey, *newObjectKey)
	if err != nil {
		utils.ExitWithError(fmt.Sprintf("Failed to rename object '%s' to '%s': %v", *oldObjectKey, *newObjectKey, err))
	}
	fmt.Printf("Successfully renamed '%s' to '%s' in '%s'.\n", *oldObjectKey, *newObjectKey, *bucketName)
}

func printUsage() {
	fmt.Println("Usage: go-cfr2 <command> [flags]")
	fmt.Println("\nCommands:")
	fmt.Println("  list      List all objects in the default R2 bucket")
	fmt.Println("            Flags:")
	fmt.Println("              -b, --bucket <name> Specify the R2 bucket name (optional)")
	fmt.Println("                                   (Defaults to DefaultBucket in config)")
	fmt.Println("\n download  Download an object from the default R2 bucket")
	fmt.Println("            Flags:")
	fmt.Println("              -b, --bucket <name> Specify the R2 bucket name (optional)")
	fmt.Println("                                   (Defaults to DefaultBucket in config)")
	fmt.Println("              -k, --key <key>      Specify the object key to download (required)")
	fmt.Println("              -o, --output <path> Specify the output file path or directory (optional)")
	fmt.Println("                                   (Defaults to current directory, filename from key)")
	fmt.Println("\n  upload    Upload a file to the default R2 bucket")
	fmt.Println("            Flags:")
	fmt.Println("              -b, --bucket <name> Specify the R2 bucket name (optional)")
	fmt.Println("                                   (Defaults to DefaultBucket in config)")
	fmt.Println("              -f, --file <path>    Specify the local file to upload (required)")
	fmt.Println("              -k, --key <key>      Specify the object key for the uploaded file (required)")
	fmt.Println("\n  delete    Delete an object from the default R2 bucket")
	fmt.Println("            Flags:")
	fmt.Println("              -b, --bucket <name> Specify the R2 bucket name (optional)")
	fmt.Println("                                   (Defaults to DefaultBucket in config)")
	fmt.Println("              -k, --key <key>      Specify the object key to delete (required)")
	fmt.Println("\n rename    Rename an object in the default R2 bucket")
	fmt.Println("            Flags:")
	fmt.Println("              -b, --bucket <name> Specify the R2 bucket name (optional)")
	fmt.Println("                                   (Defaults to DefaultBucket in config)")
	fmt.Println("              -o, --old-key <key>   Specify the old object key to rename (required)")
	fmt.Println("              -n, --new-key <key>   Specify the new object key (required)")
	fmt.Println("\n presign   Generate a presigned URL for an object with default 24-hour expiration")
	fmt.Println("            Flags:")
	fmt.Println("              -b, --bucket <name> Specify the R2 bucket name (optional)")
	fmt.Println("                                   (Defaults to DefaultBucket in config)")
	fmt.Println("              -k, --key <key>      Specify the object key (required)")
	fmt.Println("              -e, --expiry <hours> Specify the URL expiry time in hours (optional)")
	fmt.Println("                                   (Defaults to 24 hours)")
}

func handlePresignCommand(ctx context.Context, client *s3.Client, cfg *config.R2Config) {
	presignFlags := flag.NewFlagSet("presign", flag.ExitOnError)
	bucketName := presignFlags.String("b", cfg.DefaultBucket, "Specify the R2 bucket name (optional)")
	presignFlags.StringVar(bucketName, "bucket", cfg.DefaultBucket, "Specify the R2 bucket name (optional)")
	objectKey := presignFlags.String("k", "", "Specify the object key (required)")
	presignFlags.StringVar(objectKey, "key", "", "Specify the object key (required)")
	expiryHours := presignFlags.Int64("e", 24, "Specify the URL expiry time in hours (optional)")
	presignFlags.Int64Var(expiryHours, "expiry", 24, "Specify the URL expiry time in hours (optional)")
	presignFlags.Parse(os.Args[2:])

	if *bucketName == "" {
	utils.ExitWithError("Bucket name not specified. Use -b or --bucket flag, or set DefaultBucket in config.")
	}
	if *objectKey == "" {
		utils.ExitWithError("Object key not specified. Use -k or --key flag.")
	}

	fmt.Printf("Generating presigned URL for '%s' in bucket '%s' with %d-hour expiry...\n", *objectKey, *bucketName, *expiryHours)
	url, err := r2.GeneratePresignedURLWithExpiry(ctx, client, *bucketName, *objectKey, time.Duration(*expiryHours)*time.Hour)
	if err != nil {
	utils.ExitWithError(fmt.Sprintf("Failed to generate presigned URL for object '%s': %v", *objectKey, err))
	}
	fmt.Printf("Presigned URL: %s\n", url)
}
