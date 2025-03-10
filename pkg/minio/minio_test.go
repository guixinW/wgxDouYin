package minio

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRemoveBucket(t *testing.T) {
	tmpBucketName := "dc4a65747c0646c4a6eed59fb5404617"
	ctx := context.Background()
	exist, errEx := minioClient.BucketExists(ctx, tmpBucketName)
	if errEx != nil {
		t.Fatalf(errEx.Error())
	}
	if !exist {
		t.Fatalf("bucket is not exist")
	}
	err := minioClient.RemoveBucket(ctx, tmpBucketName)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestRemoveBucketsObject(t *testing.T) {
	tmpBucketName := "dc4a65747c0646c4a6eed59fb5404617"
	ctx := context.Background()
	exist, errEx := minioClient.BucketExists(context.Background(), tmpBucketName)
	if errEx != nil {
		t.Fatalf(errEx.Error())
	}
	if !exist {
		t.Fatalf("bucket is not exist")
	}
	objectCh := minioClient.ListObjects(ctx, tmpBucketName, minio.ListObjectsOptions{Recursive: true})
	removeObjects := make([]minio.ObjectInfo, 0)
	for object := range objectCh {
		removeObjects = append(removeObjects, minio.ObjectInfo{Key: object.Key})
	}
	deleteChan := make(chan minio.ObjectInfo)
	go func(rmObj []minio.ObjectInfo) {
		for _, obj := range rmObj {
			deleteChan <- obj
		}
		close(deleteChan)
	}(removeObjects)
	errorCh := minioClient.RemoveObjects(ctx, tmpBucketName, deleteChan, minio.RemoveObjectsOptions{})
	for e := range errorCh {
		t.Fatalf("删除失败：%v", e)
	}
}

func TestUploadFileByIO(t *testing.T) {
	tmpBucketName := "test"
	tmpObjectName := "testObject"
	tmpFilePath := "test.mp4"
	contentType := "application/mp4"
	ctx := context.Background()
	exist, errEx := minioClient.BucketExists(ctx, tmpBucketName)
	if exist && errEx != nil {
		err := minioClient.RemoveBucket(ctx, tmpBucketName)
		if err != nil {
			t.Error(err)
		}
	}
	err := CreateBucket(tmpBucketName)
	ExpectEqual(err, nil, t)

	r := gin.Default()
	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		ExpectEqual(err, nil, t)
		fp, err := file.Open()
		ExpectEqual(err, nil, t)
		size, err := UploadFileByIO(tmpBucketName, tmpObjectName, fp, file.Size, contentType)
		ExpectEqual(size, file.Size, t)
		ExpectEqual(err, nil, t)
	})
	go func() {
		err := r.Run("127.0.0.1:4001")
		if err != nil {
			t.Error(err.Error())
		}
	}()
	time.Sleep(5 * time.Second)
	file, err := os.Open(tmpFilePath)
	ExpectEqual(err, nil, t)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			t.Error(err.Error())
		}
	}(file)

	content := &bytes.Buffer{}
	writer := multipart.NewWriter(content)
	form, err := writer.CreateFormFile("file", filepath.Base(tmpFilePath))
	ExpectEqual(err, nil, t)
	_, err = io.Copy(form, file)
	ExpectEqual(err, nil, t)
	err = writer.Close()
	ExpectEqual(err, nil, t)
	req, err := http.NewRequest("POST", "http://127.0.0.1:4001/upload", content)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ExpectEqual(err, nil, t)
	client := &http.Client{}
	_, err = client.Do(req)
	ExpectEqual(err, nil, t)
}

func TestGetFileTemporaryURL(t *testing.T) {
	bucketName := "test"
	objectName := "testObject"
	url, err := GetFileTemporaryURL(bucketName, objectName)
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(url)
}
