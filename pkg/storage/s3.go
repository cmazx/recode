package storage

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"os"
	"strings"
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
	publicBucket := os.Getenv("S3_STORAGE_PUBLIC_BUCKET")
	opt := minio.MakeBucketOptions{Region: location}

	err = minioClient.MakeBucket(ctx, bucketName, opt)
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

	if publicBucket == "1" {
		policy := fmt.Sprintf(`{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::%s/*"],"Sid": ""}]}`,
			bucketName)
		err = minioClient.SetBucketPolicy(context.Background(), bucketName, policy)
		if err != nil {
			log.Fatalln(err)
		}
	}

	return &S3Storage{minioClient, bucketName, strings.TrimRight(publicUrl, "/")}
}

func (s S3Storage) GetObjectUrl(storagePath string) string {
	return strings.TrimRight(s.publicUrl, "/") + "/" + s.storageBucket + "/" + strings.TrimLeft(storagePath, "/")
}

func (s S3Storage) Put(localPath string, storagePath string, isPublic bool) (string, error) {
	log.Printf("Put file %s to storage path %s", localPath, storagePath)
	mime, err := GetFileContentType(localPath)
	log.Printf("Mime type %s", mime)
	if err != nil {
		return "", err
	}

	userMetaData := map[string]string{"x-amz-acl": "private"}
	if isPublic {
		userMetaData["x-amz-acl"] = "public-read"
	}
	fmt.Printf("Store object %s with metadata %v\n", storagePath, userMetaData)
	opts := minio.PutObjectOptions{ContentType: mime, UserMetadata: userMetaData}
	info, err := s.s3Client.FPutObject(context.Background(), s.storageBucket, storagePath, localPath, opts)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", storagePath, info.Size)
	return s.GetObjectUrl(storagePath), nil
}
