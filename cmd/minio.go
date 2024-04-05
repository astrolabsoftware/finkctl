package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"syscall"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func setMinioClient(c S3Config) *minio.Client {
	// TODO implement optional port-forward
	// see https://stackoverflow.com/questions/59027739/upgrading-connection-error-in-port-forwarding-via-client-go
	token := ""

	// Initialize minio client object.
	minioClient, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKeyID, c.SecretAccessKey, token),
		Secure: c.UseSSL,
	})
	if err != nil {
		slog.Error("Unable to create minio client", "error", err)
		syscall.Exit(1)
	}
	return minioClient

}

func listBucket(minioClient *minio.Client) {
	buckets, err := minioClient.ListBuckets(context.Background())
	if err != nil {
		slog.Error("Unable to list S3 bucket", "error", err)
		syscall.Exit(1)
	}
	for _, bucket := range buckets {
		fmt.Println(bucket)
	}
}

func bucketExists(minioClient *minio.Client, bucketName string) bool {
	found, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		slog.Error("Unable to check if S3 bucket exists", "error", err)
		syscall.Exit(1)
	}
	return found
}

func makeBucket(minioClient *minio.Client, bucketName string) {
	err := minioClient.MakeBucket(context.Background(), bucketName,
		minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
	if err != nil {
		slog.Error("Unable to create S3 bucket", "error", err)
		syscall.Exit(1)
	}
}
