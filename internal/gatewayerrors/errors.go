package gatewayerrors

type BankError struct {
	Err        error
	StatusCode int
}

func (be *BankError) Error() string {
	return be.Err.Error()
}

func NewBankError(err error, statusCode int) *BankError {
	return &BankError{
		Err:        err,
		StatusCode: statusCode,
	}
}

type ValidationError struct {
	Err   error
	Field string
	ID    string
}

func (ve *ValidationError) Error() string {
	return ve.Err.Error()
}

func (ve *ValidationError) GetFieldError() string {
	return ve.Field
}

func (ve *ValidationError) GetID() string {
	return ve.ID

}

func NewValidationError(err error, id, field string) *ValidationError {
	return &ValidationError{
		Err:   err,
		Field: field,
		ID:    id,
	}
}
