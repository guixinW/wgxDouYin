package tool

import ffmpeg_go "github.com/u2takey/ffmpeg-go"

func ConvertToHLS(inputFileName, outputFileName string) error {
	err := ffmpeg_go.Input(inputFileName).Output(outputFileName, ffmpeg_go.KwArgs{
		"codec:v":          "libx264",
		"codec:a":          "aac",
		"hls_time":         "2",
		"hls_list_size":    "0",
		"force_key_frames": "expr:gte(t,n_forced*2)",
		"f":                "hls",
	}).Run()
	return err
}
