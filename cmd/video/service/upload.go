package service

import (
	"bytes"
	"wgxDouYin/pkg/minio"
)

func uploadVideo(data []byte, videoTitle string) (string, error) {
	reader := bytes.NewReader(data)
	contentType := "application/mp4"
	_, err := minio.UploadFileByIO(minio.VideoBucketName, videoTitle, reader, int64(len(data)), contentType)
	if err != nil {
		logger.Errorf("视频上传至minio失败：%v", err.Error())
		return "", err
	}
	playUrl, err := minio.GetFileTemporaryURL(minio.VideoBucketName, videoTitle)
	if err != nil {
		logger.Errorf("服务器内部错误：视频获取失败：%s", err.Error())
		return "", err
	}
	logger.Infof("上传视频路径：%v", playUrl)
	return playUrl, nil
}

func VideoPublish(data []byte, videoTitle string) error {
	_, err := uploadVideo(data, videoTitle)
	if err != nil {
		return err
	}
	return nil
}
