package tool

import "testing"

func TestConvertToHLS(t *testing.T) {
	inputFile := "test.mp4"
	outputFile := "output.m3u8"
	if err := ConvertToHLS(inputFile, outputFile); err != nil {
		t.Fatalf("failed to convert to hls: %v", err)
	}
}
