package minio

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"wgxDouYin/pkg/viper"
)

var (
	minioClient     *minio.Client
	minioConfig     = viper.Init("minio")
	EndPoint        = minioConfig.Viper.GetString("minio.Endpoint")
	AccessKeyId     = minioConfig.Viper.GetString("minio.AccessKeyId")
	SecretAccessKey = minioConfig.Viper.GetString("minio.SecretAccessKey")
	UseSSL          = minioConfig.Viper.GetBool("minio.UseSSL")
	VideoBucketName = minioConfig.Viper.GetString("minio.VideoBucketName")
	ExpireTime      = minioConfig.Viper.GetUint32("minio.ExpireTime")
)

func init() {
	s3client, err := minio.New(EndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(AccessKeyId, SecretAccessKey, ""),
		Secure: UseSSL,
	})

	if err != nil {
		panic(err)
	}
	minioClient = s3client
	if err := CreateBucket(VideoBucketName); err != nil {
		panic(err)
	}
}
