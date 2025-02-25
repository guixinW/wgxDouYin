package minio

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUploadFileByIO(t *testing.T) {
	tmpBucketName := "dc4a65747c0646c4a6eed59fb5404617"
	tmpObjectName := "aasdjals923ijnsjnfao3i"
	tmpFilePath := "fileexist.mp4"
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
		err = minioClient.RemoveBucket(ctx, tmpBucketName)
		if err != nil {
			t.Error(err)
		}
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
