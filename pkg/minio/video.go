package minio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func mp4ToByte(fileName string) []byte {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return nil
	}
	file, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return nil
	}
	return buf.Bytes()
}

func saveToTempFile(videoId uint64, data []byte) (string, error) {
	tempFile, err := os.CreateTemp("", fmt.Sprintf("video_%v.mp4", videoId))
	if err != nil {
		return "", err
	}
	defer tempFile.Close()
	_, err = tempFile.Write(data)
	if err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}

func convertToHLS(inputFile string, outputDir string) error {
	outputM3U8 := filepath.Join(outputDir, "output.m3u8")
	cmd := exec.Command("ffmpeg",
		"-i", inputFile, // 输入文件
		"-profile:v", "baseline", // 编码配置
		"-level", "3.0", // H.264 编码等级
		"-s", "1280x720", // 设置分辨率
		"-hls_time", "120", // 每个 .ts 文件的时长
		"-hls_list_size", "0", // 无限个片段
		"-f", "hls", // 强制输出为 HLS 格式
		outputM3U8, // 输出的播放列表文件
	)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running FFmpeg: %w", err)
	}
	return nil
}
