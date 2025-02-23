package genericerrors

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
