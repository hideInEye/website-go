package errors

// HTTPError carries the public status and message for a business error.
type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func NewHTTP(status int, message string) error {
	return &HTTPError{Status: status, Message: message}
}
