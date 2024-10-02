package custom_errors

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func NewNotFound(message string) *NotFoundError {
	return &NotFoundError{Message: message}
}
