package datatrans

import "fmt"

type ErrorResponse struct {
	HTTPStatusCode int
	ErrorDetail    ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (s ErrorResponse) Error() string {
	if s.ErrorDetail.Code == "" {
		return ""
	}
	return fmt.Sprintf(
		"HTTPStatusCode:%d Code:%q, Message:%q",
		s.HTTPStatusCode,
		s.ErrorDetail.Code,
		s.ErrorDetail.Message,
	)
}
