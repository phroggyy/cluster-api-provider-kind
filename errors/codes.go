package errors

const DeleteFailedErr = "delete_failed"
const CreateFailedErr = "create_failed"

type CodedError struct {
	Code string
	Err  error
}

func (e *CodedError) Error() string {
	return e.Err.Error()
}

func Code(err error, code string) error {
	if err == nil {
		return nil
	}

	return &CodedError{
		Code: code,
		Err:  err,
	}
}
