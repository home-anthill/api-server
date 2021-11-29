package errors

type Status int

const (
	ConnectionError Status = iota + 1
	BadParams
	NotFound
)

type SendGrpcError struct {
	Status  Status
	Message string
}

func (se SendGrpcError) Error() string {
	return se.Message
}
