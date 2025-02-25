package minio

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	"io"
	"runtime/debug"
	"testing"
)

func ExpectEqual(left interface{}, right interface{}, t *testing.T) {
	if left != right {
		t.Fatalf("expected %v == %v\n%s", left, right, debug.Stack())
	}
}

func CreateBucket(bucketName string) error {
	if len(bucketName) <= 0 {
		return errors.New("bucketName invalid")
	}
	ctx := context.Background()
	if err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
		exists, errEx := minioClient.BucketExists(ctx, bucketName)
		if exists && errEx != nil {
		} else {
			return errEx
		}
	}
	return nil
}

func UploadFileByPath(bucketName, objectName, path, contentType string) (int64, error) {
	if len(bucketName) <= 0 || len(objectName) <= 0 || len(path) <= 0 {
		return -1, errors.New("invalid argument")
	}

	uploadInfo, err := minioClient.FPutObject(context.Background(), bucketName, objectName, path, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		return -1, err
	}

	return uploadInfo.Size, nil
}

func UploadFileByIO(bucketName, objectName string, reader io.Reader, size int64, contentType string) (int64, error) {
	if len(bucketName) <= 0 || len(objectName) <= 0 {
		return -1, errors.New("invalid argument")
	}
	uploadInfo, err := minioClient.PutObject(context.Background(), bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return -1, err
	}
	return uploadInfo.Size, nil
}
