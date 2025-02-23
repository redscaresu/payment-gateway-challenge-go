package gatewayerrors

/*
Pretty much what it says on the tin, here I created some custom errors for our service so that we could create specific types that we could check against in the handler and also keep some additional info.

For example arguable YAGNI but I included the field as part of the validation errors,I am really suprised the spec did not mention the potential for passing back the field error to the customer.  We need to have a chat with product management and have a bit more of a think how we pass back errors to the customers I think, for the timebeing we log the field that the customer had an error on to help in troubleshooting in case they come and contact us.
*/

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
