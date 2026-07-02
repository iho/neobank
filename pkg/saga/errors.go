package saga

import "fmt"

// BusinessError represents an expected domain failure (no compensation needed).
type BusinessError struct {
	Code    string
	Message string
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{Code: code, Message: message}
}