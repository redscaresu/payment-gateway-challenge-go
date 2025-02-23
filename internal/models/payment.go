package models

/*

If I had more time I would completely split out the models used in the handlers from the models used throughout the program.  Because I dont like the presentation tier being tied to implementation, for example in the PostPayment handler I am just reusing PostPaymentResponse for the happy path and possible a new validation error.

*/

type PostPaymentHandlerRequest struct {
	CardNumber  int    `json:"card_number"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	Currency    string `json:"currency"`
	Amount      int    `json:"amount"`
	Cvv         int    `json:"cvv"`
}

type GetPaymentHandlerResponse struct {
	Id                 string `json:"id"`
	Status             string `json:"status"`
	LastFourCardDigits int    `json:"last_four_card_digits"`
	ExpiryMonth        int    `json:"expiry_month"`
	ExpiryYear         int    `json:"expiry_year"`
	Currency           string `json:"currency"`
	Amount             int    `json:"amount"`
}

type PostPaymentRequest struct {
	CardNumberLastFour int    `json:"card_number_last_four"`
	ExpiryMonth        int    `json:"expiry_month"`
	ExpiryYear         int    `json:"expiry_year"`
	Currency           string `json:"currency"`
	Amount             int    `json:"amount"`
	Cvv                int    `json:"cvv"`
}

type PostPaymentResponse struct {
	Id                 string `json:"id"`
	PaymentStatus      string `json:"payment_status"`
	CardNumberLastFour int    `json:"card_number_last_four"`
	ExpiryMonth        int    `json:"expiry_month"`
	ExpiryYear         int    `json:"expiry_year"`
	Currency           string `json:"currency"`
	Amount             int    `json:"amount"`
}

type GetPaymentResponse struct {
	Id                 string `json:"id"`
	PaymentStatus      string `json:"payment_status"`
	CardNumberLastFour int    `json:"card_number_last_four"`
	ExpiryMonth        int    `json:"expiry_month"`
	ExpiryYear         int    `json:"expiry_year"`
	Currency           string `json:"currency"`
	Amount             int    `json:"amount"`
}

type PostPaymentBankRequest struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	Currency   string `json:"currency"`
	Amount     int    `json:"amount"`
	CVV        string `json:"cvv"`
}

type PostPaymentBankResponse struct {
	Authorised        bool   `json:"authorized"`
	AuthorizationCode string `json:"authorization_code"`
}

type PostPayment400Response struct {
	Id            string `json:"id"`
	PaymentStatus string `json:"payment_status"`
}
