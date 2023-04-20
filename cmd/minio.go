package cmd

import (
	"context"
	"fmt"

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
		logger.Fatalf("Unable to create minio client: %s", err)
	}

	logger.Debugf("%#v\n", minioClient)
	return minioClient

}

func listBucket(minioClient *minio.Client) {
	buckets, err := minioClient.ListBuckets(context.Background())
	if err != nil {
		logger.Fatal(err)
	}
	for _, bucket := range buckets {
		fmt.Println(bucket)
	}
}

func makeBucket(minioClient *minio.Client, bucketName string) {
	err := minioClient.MakeBucket(context.Background(), bucketName,
		minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
	if err != nil {
		logger.Fatal(err)
	}
}
