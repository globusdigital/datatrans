package datatrans

import "time"

type ResponseAuthorize struct {
	AcquirerAuthorizationCode string `json:"acquirerAuthorizationCode"`
}

type RequestValidateAlias struct {
	Currency   string     `json:"currency"`
	RefNo      string     `json:"refno"`
	CardSimple *CardAlias `json:"card,omitempty"`
	//autoSettle
	RefNo2 string `json:"refno2,omitempty"`
	// TODO add more fields
}

type RequestSettle struct {
	Amount     int                    `json:"amount"`
	Currency   string                 `json:"currency"`
	RefNo      string                 `json:"refno"`
	RefNo2     string                 `json:"refno2,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	// TODO add more fields
}

type RequestCredit struct {
	Amount     int                    `json:"amount"`
	Currency   string                 `json:"currency"`
	RefNo      string                 `json:"refno"`
	RefNo2     string                 `json:"refno2,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	// TODO add more fields
}

type ResponseCardMasked struct {
	TransactionId             string            `json:"transactionId"`
	AcquirerAuthorizationCode string            `json:"acquirerAuthorizationCode"`
	Card                      *CardMaskedSimple `json:"card,omitempty"` // only set in case of CreditAuthorize
}

type CardMaskedSimple struct {
	Masked string `json:"masked"`
}

type RequestCreditAuthorize struct {
	Currency   string     `json:"currency"`
	RefNo      string     `json:"refno"`
	Card       *CardAlias `json:"card,omitempty"`
	Amount     int        `json:"amount"`
	AutoSettle bool       `json:"autoSettle,omitempty"`
	Refno2     string     `json:"refno2,omitempty"`
	// TODO add more fields
}

type CardAlias struct {
	Alias       string `json:"alias"`
	ExpiryMonth string `json:"expiryMonth"`
	ExpiryYear  string `json:"expiryYear"`
}

type ResponseStatus struct {
	TransactionID string `json:"transactionId"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	Currency      string `json:"currency"`
	RefNo         string `json:"refno"`
	PaymentMethod string `json:"paymentMethod"`
	Detail        struct {
		Init struct {
			Expires time.Time `json:"expires"`
		} `json:"init"`
		Authorize struct {
			Amount                    int    `json:"amount"`
			AcquirerAuthorizationCode string `json:"acquirerAuthorizationCode"`
		} `json:"authorize"`
		// TODO add more fields https://api-reference.datatrans.ch/#operation/status
	} `json:"detail"`
	Card  *CardExtended `json:"card"`
	Twint *struct {
		Alias string `json:"alias"`
	} `json:"twi"`
	// TODO add more fields https://api-reference.datatrans.ch/#operation/status

	History []History `json:"history"`
}

type CardExtended struct {
	Masked      string `json:"masked"`
	ExpiryMonth string `json:"expiryMonth"`
	ExpiryYear  string `json:"expiryYear"`
	Info        struct {
		Brand   string `json:"brand"`
		Type    string `json:"type"`
		Usage   string `json:"usage"`
		Country string `json:"country"`
		Issuer  string `json:"issuer"`
	} `json:"info"`
}

type History struct {
	Action  string    `json:"action"`
	Amount  int       `json:"amount"`
	Source  string    `json:"source"`
	Date    time.Time `json:"date"`
	Success bool      `json:"success"`
	IP      string    `json:"ip"`
}
