package customerrors

// Status int
type Status int

// ConnectionError and BadParams enums
const (
	ConnectionError Status = iota + 1
	BadParams
)

// GrpcSendError struct
type GrpcSendError struct {
	Status  Status
	Message string
}

// Error function
func (se GrpcSendError) Error() string {
	return se.Message
}
