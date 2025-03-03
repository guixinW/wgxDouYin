package response

type Base struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

func ErrorResponse(errorMsg string) Base {
	return Base{
		StatusCode: -1,
		StatusMsg:  errorMsg,
	}
}
