package errors

type Status int

const (
	ConnectionError Status = iota + 1
	BadParams
)

type GrpcSendError struct {
	Status  Status
	Message string
}

func (se GrpcSendError) Error() string {
	return se.Message
}
