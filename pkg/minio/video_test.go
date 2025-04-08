package minio

import "testing"

func TestTransferHLS(t *testing.T) {
	videoByte := mp4ToByte("生活大爆炸.mp4")
	fileName, err := saveToTempFile(1, videoByte)
	if err != nil {
		t.Fatalf("failed to save temp file:%v", err)
	}
	err = convertToHLS(fileName, "")
	if err != nil {
		t.Fatalf("failed to convert to hls file: %v", err)
	}
}
