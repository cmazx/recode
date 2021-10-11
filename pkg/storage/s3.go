package storage

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"os"
)

type S3Storage struct {
	s3Client      *minio.Client
	storageBucket string
	publicUrl     string
}

func NewS3Storage() Storage {

	// Initialize minio client object.
	minioClient, err := minio.New(os.Getenv("S3_STORAGE_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("S3_STORAGE_ACCESS_KEY"), os.Getenv("S3_STORAGE_SECRET_KEY"), ""),
		Secure: os.Getenv("S3_STORAGE_USE_SSL") == "1",
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	bucketName := os.Getenv("S3_STORAGE_BUCKET_NAME")
	location := os.Getenv("S3_STORAGE_REGION")
	publicUrl := os.Getenv("S3_STORAGE_PUBLIC_URL")
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("S3 Bucket exist %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("S3 Bucket created %s\n", bucketName)
	}

	log.Printf("%#v\n", minioClient)
	var s Storage
	s = &S3Storage{minioClient, bucketName, publicUrl}
	return s
}

func (s S3Storage) Put(localPath string, storagePath string) (string, error) {
	log.Printf("Put file %s to storage path %s", localPath, storagePath)
	mime, err := GetFileContentType(storagePath)
	if err != nil {
		return "", err
	}
	opts := minio.PutObjectOptions{ContentType: mime}
	info, err := s.s3Client.FPutObject(context.Background(), s.storageBucket, storagePath, localPath, opts)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", storagePath, info.Size)
	return s.publicUrl + storagePath, nil
}
