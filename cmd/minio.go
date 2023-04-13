package cmd

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func setMinioClient() *minio.Client {
	// TODO implement optional port-forward
	// see https://stackoverflow.com/questions/59027739/upgrading-connection-error-in-port-forwarding-via-client-go
	// endpoint := "minio.minio-dev:9000"
	endpoint := "localhost:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false
	token := ""

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, token),
		Secure: useSSL,
	})
	if err != nil {
		logger.Fatal(err)
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

func makeBucket(minioClient *minio.Client) {
	err := minioClient.MakeBucket(context.Background(), "fink-broker-online",
		minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
	if err != nil {
		logger.Fatal(err)
	}
}
