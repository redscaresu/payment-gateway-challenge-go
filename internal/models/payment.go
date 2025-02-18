package models

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
	CardNumber int `json:"card_number"`
	ExpiryDate int `json:"expiry_date"`
	Currency   int `json:"currency"`
	Amount     int `json:"amount"`
	CVV        int `json:"cvv"`
}

type PostPaymentBankResponse struct {
	Authorised bool   `json:"authorised"`
	Message    string `json:"message"`
}
