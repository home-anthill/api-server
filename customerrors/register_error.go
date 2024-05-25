package customerrors

// ErrorWrapper struct
type ErrorWrapper struct {
	Message string `json:"message"`
	Code    int    `json:"errCode"`
	Err     error  `json:"-"`
}

// Error function
func (err ErrorWrapper) Error() string {
	if err.Err != nil {
		return err.Err.Error()
	}
	return err.Message
}

// Unwrap function
func (err ErrorWrapper) Unwrap() error {
	return err.Err
}

// Wrap function
func Wrap(code int, err error, message string) error {
	return ErrorWrapper{
		Message: message,
		Code:    code,
		Err:     err,
	}
}
