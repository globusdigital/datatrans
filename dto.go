package datatrans

import (
	"time"
)

// More fields can be added to any of the structs if needed. Just send a PR.

type customFieldsGetter interface {
	getCustomFields() map[string]interface{}
}

// CustomFields allows to extend any input with merchant specific settings.
type CustomFields map[string]interface{}

func (cf CustomFields) getCustomFields() map[string]interface{} { return cf }

type rawJSONBodySetter interface {
	setJSONRawBody([]byte)
}

// RawJSONBody includes the original response from the datatrans server. There
// might be custom fields in the response which are not included in the structs
// in this package. This type allows for unmarshaling into custom structs.
type RawJSONBody []byte

func (b *RawJSONBody) setJSONRawBody(p []byte) {
	*b = p
}

// https://api-reference.datatrans.ch/#operation/secureFieldsInit
type RequestSecureFieldsInit struct {
	Currency     string `json:"currency"`
	Amount       int    `json:"amount,omitempty"`
	ReturnUrl    string `json:"returnUrl"`
	CustomFields `json:"-"`
}

// https://api-reference.datatrans.ch/#operation/secure-fields-update
type RequestSecureFieldsUpdate struct {
	Currency     string `json:"currency"`
	Amount       int    `json:"amount,omitempty"`
	CustomFields `json:"-"`
}

// https://api-reference.datatrans.ch/#operation/init
type RequestInitialize struct {
	Currency       string            `json:"currency"`
	RefNo          string            `json:"refno"`
	RefNo2         string            `json:"refno2,omitempty"`
	AutoSettle     bool              `json:"autoSettle,omitempty"`
	Customer       *Customer         `json:"customer,omitempty"`
	Card           *Card             `json:"card,omitempty"`
	Amount         int               `json:"amount,omitempty"`
	Language       string            `json:"language,omitempty"` // Enum: "en" "de" "fr" "it" "es" "el" "no" "da" "pl" "pt" "ru" "ja"
	PaymentMethods []string          `json:"paymentMethods,omitempty"`
	Theme          *Theme            `json:"theme,omitempty"`
	Redirect       *Redirect         `json:"redirect,omitempty"`
	Option         *InitializeOption `json:"option,omitempty"`
	CustomFields   `json:"-"`
}

type ResponseInitialize struct {
	Location      string `json:"location,omitempty"` // A URL where the users browser needs to be redirect to complete the payment. This redirect is only needed when using Redirect Mode. For Lightbox Mode the returned transactionId can be used to start the payment page.
	TransactionId string `json:"transactionId,omitempty"`
	MobileToken   string `json:"mobileToken,omitempty"`
	RawJSONBody   `json:"raw,omitempty"`
}

type RequestAuthorize struct {
	Amount     int    `json:"amount,omitempty"`
	Currency   string `json:"currency,omitempty"`
	RefNo      string `json:"refno,omitempty"`
	RefNo2     string `json:"refno2,omitempty"`
	AutoSettle bool   `json:"autoSettle,omitempty"`
	// The card object to be submitted when authorizing with an existing credit
	// card alias.
	Card         *Card `json:"card,omitempty"`
	CustomFields `json:"-"`
}

type ResponseAuthorize struct {
	AcquirerAuthorizationCode string `json:"acquirerAuthorizationCode"`
	RawJSONBody               `json:"raw,omitempty"`
}

type RequestAuthorizeTransaction struct {
	RefNo        string `json:"refno,omitempty"`
	Amount       int    `json:"amount,omitempty"`
	AutoSettle   bool   `json:"autoSettle,omitempty"`
	RefNo2       string `json:"refno2,omitempty"`
	CustomFields `json:"-"`
}

type RequestValidateAlias struct {
	Currency     string `json:"currency,omitempty"`
	RefNo        string `json:"refno,omitempty"`
	RefNo2       string `json:"refno2,omitempty"`
	Card         *Card  `json:"card,omitempty"`
	CustomFields `json:"-"`
}

type RequestSettle struct {
	Amount       int    `json:"amount,omitempty"`
	Currency     string `json:"currency,omitempty"`
	RefNo        string `json:"refno,omitempty"`
	RefNo2       string `json:"refno2,omitempty"`
	CustomFields `json:"-"`
}

type RequestCredit struct {
	Amount       int    `json:"amount,omitempty"`
	Currency     string `json:"currency,omitempty"`
	RefNo        string `json:"refno,omitempty"`
	RefNo2       string `json:"refno2,omitempty"`
	CustomFields `json:"-"`
}

type RequestCreditAuthorize struct {
	Currency     string `json:"currency,omitempty"`
	RefNo        string `json:"refno,omitempty"`
	Card         *Card  `json:"card,omitempty"`
	Amount       int    `json:"amount,omitempty"`
	AutoSettle   bool   `json:"autoSettle,omitempty"`
	Refno2       string `json:"refno2,omitempty"`
	CustomFields `json:"-"`
}

type ResponseCardMasked struct {
	TransactionId             string            `json:"transactionId,omitempty"`
	AcquirerAuthorizationCode string            `json:"acquirerAuthorizationCode,omitempty"`
	Card                      *CardMaskedSimple `json:"card,omitempty"` // only set in case of CreditAuthorize
	RawJSONBody               `json:"raw,omitempty"`
}

type CardMaskedSimple struct {
	Masked string `json:"masked,omitempty"`
}

type Card struct {
	Alias       string `json:"alias,omitempty"`
	AliasCVV    string `json:"aliasCVV,omitempty"`
	ExpiryMonth string `json:"expiryMonth,omitempty"`
	ExpiryYear  string `json:"expiryYear,omitempty"`
	ThreeD      ThreeD `json:"3D,omitempty"`
}
type ThreeDSRequestorAuthenticationInfo struct {
	ThreeDSReqAuthMethod    string `json:"threeDSReqAuthMethod,omitempty"`
	ThreeDSReqAuthTimestamp string `json:"threeDSReqAuthTimestamp,omitempty"`
	ThreeDSReqAuthData      string `json:"threeDSReqAuthData,omitempty"`
}
type ThreeDSRequestorPriorAuthenticationInfo struct {
	ThreeDSReqPriorRef           string `json:"threeDSReqPriorRef,omitempty"`
	ThreeDSReqPriorAuthMethod    string `json:"threeDSReqPriorAuthMethod,omitempty"`
	ThreeDSReqPriorAuthTimestamp string `json:"threeDSReqPriorAuthTimestamp,omitempty"`
	ThreeDSReqPriorAuthData      string `json:"threeDSReqPriorAuthData,omitempty"`
}
type ThreeDSRequestor struct {
	ThreeDSRequestorAuthenticationInd       string                                  `json:"threeDSRequestorAuthenticationInd,omitempty"`
	ThreeDSRequestorAuthenticationInfo      ThreeDSRequestorAuthenticationInfo      `json:"threeDSRequestorAuthenticationInfo,omitempty"`
	ThreeDSRequestorChallengeInd            string                                  `json:"threeDSRequestorChallengeInd,omitempty"`
	ThreeDSRequestorPriorAuthenticationInfo ThreeDSRequestorPriorAuthenticationInfo `json:"threeDSRequestorPriorAuthenticationInfo,omitempty"`
}
type AcctInfo struct {
	ChAccDate             string `json:"chAccDate,omitempty"`
	ChAccChangeInd        string `json:"chAccChangeInd,omitempty"`
	ChAccChange           string `json:"chAccChange,omitempty"`
	ChAccPwChangeInd      string `json:"chAccPwChangeInd,omitempty"`
	ChAccPwChange         string `json:"chAccPwChange,omitempty"`
	ShipAddressUsageInd   string `json:"shipAddressUsageInd,omitempty"`
	ShipAddressUsage      string `json:"shipAddressUsage,omitempty"`
	TxnActivityDay        int    `json:"txnActivityDay,omitempty"`
	TxnActivityYear       int    `json:"txnActivityYear,omitempty"`
	ProvisionAttemptsDay  int    `json:"provisionAttemptsDay,omitempty"`
	NbPurchaseAccount     int    `json:"nbPurchaseAccount,omitempty"`
	SuspiciousAccActivity string `json:"suspiciousAccActivity,omitempty"`
	ShipNameIndicator     string `json:"shipNameIndicator,omitempty"`
	PaymentAccInd         string `json:"paymentAccInd,omitempty"`
	PaymentAccAge         string `json:"paymentAccAge,omitempty"`
}
type CardholderAccount struct {
	AcctType       string   `json:"acctType,omitempty"`
	CardExpiryDate string   `json:"cardExpiryDate,omitempty"`
	AcctInfo       AcctInfo `json:"acctInfo,omitempty"`
	AcctNumber     string   `json:"acctNumber,omitempty"`
	SchemeID       string   `json:"schemeId,omitempty"`
	AcctID         string   `json:"acctID,omitempty"`
	PayTokenInd    bool     `json:"payTokenInd,omitempty"`
}
type HomePhone struct {
	Cc         string `json:"cc,omitempty"`
	Subscriber string `json:"subscriber,omitempty"`
}
type MobilePhone struct {
	Cc         string `json:"cc,omitempty"`
	Subscriber string `json:"subscriber,omitempty"`
}
type WorkPhone struct {
	Cc         string `json:"cc,omitempty"`
	Subscriber string `json:"subscriber,omitempty"`
}
type Cardholder struct {
	AddrMatch        string       `json:"addrMatch,omitempty"`
	BillAddrCity     string       `json:"billAddrCity,omitempty"`
	BillAddrCountry  string       `json:"billAddrCountry,omitempty"`
	BillAddrLine1    string       `json:"billAddrLine1,omitempty"`
	BillAddrLine2    string       `json:"billAddrLine2,omitempty"`
	BillAddrLine3    string       `json:"billAddrLine3,omitempty"`
	BillAddrPostCode string       `json:"billAddrPostCode,omitempty"`
	BillAddrState    string       `json:"billAddrState,omitempty"`
	Email            string       `json:"email,omitempty"`
	HomePhone        *HomePhone   `json:"homePhone,omitempty"`
	MobilePhone      *MobilePhone `json:"mobilePhone,omitempty"`
	WorkPhone        *WorkPhone   `json:"workPhone,omitempty"`
	CardholderName   string       `json:"cardholderName,omitempty"`
	ShipAddrCity     string       `json:"shipAddrCity,omitempty"`
	ShipAddrCountry  string       `json:"shipAddrCountry,omitempty"`
	ShipAddrLine1    string       `json:"shipAddrLine1,omitempty"`
	ShipAddrLine2    string       `json:"shipAddrLine2,omitempty"`
	ShipAddrLine3    string       `json:"shipAddrLine3,omitempty"`
	ShipAddrPostCode string       `json:"shipAddrPostCode,omitempty"`
	ShipAddrState    string       `json:"shipAddrState,omitempty"`
}
type MerchantRiskIndicator struct {
	ShipIndicator        string `json:"shipIndicator,omitempty"`
	DeliveryTimeframe    string `json:"deliveryTimeframe,omitempty"`
	DeliveryEmailAddress string `json:"deliveryEmailAddress,omitempty"`
	ReorderItemsInd      string `json:"reorderItemsInd,omitempty"`
	PreOrderPurchaseInd  string `json:"preOrderPurchaseInd,omitempty"`
	PreOrderDate         string `json:"preOrderDate,omitempty"`
	GiftCardAmount       int    `json:"giftCardAmount,omitempty"`
	GiftCardCurr         string `json:"giftCardCurr,omitempty"`
	GiftCardCount        string `json:"giftCardCount,omitempty"`
}
type Purchase struct {
	PurchaseInstalData    int                   `json:"purchaseInstalData,omitempty"`
	MerchantRiskIndicator MerchantRiskIndicator `json:"merchantRiskIndicator,omitempty"`
	PurchaseAmount        int                   `json:"purchaseAmount,omitempty"`
	PurchaseCurrency      string                `json:"purchaseCurrency,omitempty"`
	PurchaseExponent      int                   `json:"purchaseExponent,omitempty"`
	PurchaseDate          string                `json:"purchaseDate,omitempty"`
	RecurringExpiry       string                `json:"recurringExpiry,omitempty"`
	RecurringFrequency    int                   `json:"recurringFrequency,omitempty"`
	TransType             string                `json:"transType,omitempty"`
}
type Acquirer struct {
	AcquirerBin        string `json:"acquirerBin,omitempty"`
	AcquirerMerchantID string `json:"acquirerMerchantId,omitempty"`
}
type Merchant struct {
	Mcc                 string `json:"mcc,omitempty"`
	MerchantCountryCode string `json:"merchantCountryCode,omitempty"`
	MerchantName        string `json:"merchantName,omitempty"`
}
type BroadInfo struct {
	Message string `json:"message,omitempty"`
}
type DeviceRenderOptions struct {
	SdkInterface string   `json:"sdkInterface,omitempty"`
	SdkUIType    []string `json:"sdkUiType,omitempty"`
}
type Data struct {
	ValueOne string `json:"valueOne,omitempty"`
	ValueTwo string `json:"valueTwo,omitempty"`
}
type MessageExtension struct {
	ID                   string `json:"id,omitempty"`
	Name                 string `json:"name,omitempty"`
	CriticalityIndicator bool   `json:"criticalityIndicator,omitempty"`
	Data                 Data   `json:"data,omitempty"`
}
type BrowserInformation struct {
	BrowserAcceptHeader string `json:"browserAcceptHeader,omitempty"`
	BrowserIP           string `json:"browserIP,omitempty"`
	BrowserJavaEnabled  bool   `json:"browserJavaEnabled,omitempty"`
	BrowserLanguage     string `json:"browserLanguage,omitempty"`
	BrowserColorDepth   string `json:"browserColorDepth,omitempty"`
	BrowserScreenHeight int    `json:"browserScreenHeight,omitempty"`
	BrowserScreenWidth  int    `json:"browserScreenWidth,omitempty"`
	BrowserTZ           int    `json:"browserTZ,omitempty"`
	BrowserUserAgent    string `json:"browserUserAgent,omitempty"`
	ChallengeWindowSize string `json:"challengeWindowSize,omitempty"`
}
type ThreeD struct {
	PreferredProtocolVersion        string               `json:"preferredProtocolVersion,omitempty"`
	EnforcePreferredProtocolVersion bool                 `json:"enforcePreferredProtocolVersion,omitempty"`
	DeviceChannel                   string               `json:"deviceChannel,omitempty"`
	MessageCategory                 string               `json:"messageCategory,omitempty"`
	ThreeDSCompInd                  string               `json:"threeDSCompInd,omitempty"`
	ThreeDSRequestor                *ThreeDSRequestor    `json:"threeDSRequestor,omitempty"`
	ThreeDSServerTransID            string               `json:"threeDSServerTransID,omitempty"`
	CardholderAccount               *CardholderAccount   `json:"cardholderAccount,omitempty"`
	Cardholder                      *Cardholder          `json:"cardholder,omitempty"`
	RelaxRegionalValidationRules    bool                 `json:"relaxRegionalValidationRules,omitempty"`
	Purchase                        *Purchase            `json:"purchase,omitempty"`
	Acquirer                        *Acquirer            `json:"acquirer,omitempty"`
	Merchant                        *Merchant            `json:"merchant,omitempty"`
	BroadInfo                       *BroadInfo           `json:"broadInfo,omitempty"`
	DeviceRenderOptions             *DeviceRenderOptions `json:"deviceRenderOptions,omitempty"`
	MessageExtension                []MessageExtension   `json:"messageExtension,omitempty"`
	BrowserInformation              *BrowserInformation  `json:"browserInformation,omitempty"`
	ThreeRIInd                      string               `json:"threeRIInd,omitempty"`
}

type ResponseStatus struct {
	TransactionID string `json:"transactionId,omitempty"`
	MerchantID    string `json:"merchantId,omitempty"`
	Type          string `json:"type,omitempty"`
	Status        string `json:"status,omitempty"`
	Currency      string `json:"currency,omitempty"`
	RefNo         string `json:"refno,omitempty"`
	PaymentMethod string `json:"paymentMethod,omitempty"`
	Detail        struct {
		Init struct {
			Expires time.Time `json:"expires,omitempty"` // Tells when the initialized transaction will expire if not continued - 30 minutes after initialization.
		} `json:"init,omitempty"`
		Authorize struct {
			Amount                    int    `json:"amount,omitempty"`
			AcquirerAuthorizationCode string `json:"acquirerAuthorizationCode,omitempty"`
		} `json:"authorize,omitempty"`
		Settle struct {
			Amount int `json:"amount,omitempty"`
		} `json:"settle,omitempty"`
		Credit struct {
			Amount int `json:"amount,omitempty"`
		} `json:"credit,omitempty"`
		Cancel struct {
			Reversal bool `json:"reversal,omitempty"` // Whether the transaction was reversed on acquirer side.
		} `json:"cancel,omitempty"`
		Fail struct {
			Reason  string `json:"reason,omitempty"`
			Message string `json:"message,omitempty"`
		} `json:"fail,omitempty"`
	} `json:"detail,omitempty"`
	Customer    *Customer     `json:"customer,omitempty"`
	Card        *CardExtended `json:"card,omitempty"`
	Language    string        `json:"language,omitempty"`
	History     []History     `json:"history,omitempty"`
	RawJSONBody `json:"raw,omitempty"`
}

type CardExtended struct {
	Alias           string            `json:"alias,omitempty"`
	AliasCVV        string            `json:"aliasCVV,omitempty"`
	Masked          string            `json:"masked,omitempty"`
	ExpiryMonth     string            `json:"expiryMonth,omitempty"`
	ExpiryYear      string            `json:"expiryYear,omitempty"`
	Info            *CardExtendedInfo `json:"info,omitempty"`
	WalletIndicator string            `json:"walletIndicator,omitempty"`
}

type CardExtendedInfo struct {
	Brand   string `json:"brand,omitempty"`
	Type    string `json:"type,omitempty"`
	Usage   string `json:"usage,omitempty"`
	Country string `json:"country,omitempty"`
	Issuer  string `json:"issuer,omitempty"`
}

type History struct {
	Action  string    `json:"action,omitempty"`
	Amount  int       `json:"amount,omitempty"`
	Source  string    `json:"source,omitempty"`
	Date    time.Time `json:"date,omitempty"`
	Success bool      `json:"success,omitempty"`
	IP      string    `json:"ip,omitempty"`
}

type Customer struct {
	ID                    string `json:"id,omitempty"`                    // Unique customer identifier
	Title                 string `json:"title,omitempty"`                 // Something like Ms or Mrs
	FirstName             string `json:"firstName,omitempty"`             // The first name of the customer.
	LastName              string `json:"lastName,omitempty"`              // The last name of the customer.
	Street                string `json:"street,omitempty"`                // The street of the customer.
	Street2               string `json:"street2,omitempty"`               // Additional street information. For example: '3rd floor'
	City                  string `json:"city,omitempty"`                  // The city of the customer.
	Country               string `json:"country,omitempty"`               // 2 letter ISO 3166-1 alpha-2 country code
	ZipCode               string `json:"zipCode,omitempty"`               // Zip code of the customer.
	Phone                 string `json:"phone,omitempty"`                 // Phone number of the customer.
	CellPhone             string `json:"cellPhone,omitempty"`             // Cell Phone number of the customer.
	Email                 string `json:"email,omitempty"`                 // The email address of the customer.
	Gender                string `json:"gender,omitempty"`                // Gender of the customer. female or male.
	BirthDate             string `json:"birthDate,omitempty"`             // The birth date of the customer. Must be in ISO-8601 format (YYYY-MM-DD).
	Language              string `json:"language,omitempty"`              // The language of the customer.
	Type                  string `json:"type,omitempty"`                  // P or C depending on whether the customer is private or a company. If C, the fields name and companyRegisterNumber are required
	Name                  string `json:"name,omitempty"`                  // The name of the company. Only applicable if type=C
	CompanyLegalForm      string `json:"companyLegalForm,omitempty"`      // The legal form of the company (AG, GmbH, ...)
	CompanyRegisterNumber string `json:"companyRegisterNumber,omitempty"` // The register number of the company. Only applicable if type=C
	IpAddress             string `json:"ipAddress,omitempty"`             // The ip address of the customer.
}

type Theme struct {
	// 	Theme configuration options when using the default DT2015 theme
	Name          string             `json:"name,omitempty"` // Theme name, e.g. DT2015
	Configuration ThemeConfiguration `json:"configuration,omitempty"`
}

type ThemeConfiguration struct {
	BrandColor         string `json:"brandColor,omitempty"`         // Hex notation of a color
	TextColor          string `json:"textColor,omitempty"`          // Enum: "white" "black"	The color of the text in the header bar if no logo is given
	LogoType           string `json:"logoType,omitempty"`           // Enum: "circle" "rectangle" "none" 	The header logo's display style
	LogoBorderColor    string `json:"logoBorderColor,omitempty"`    // Decides whether the logo shall be styled with a border around it, if the value is true the default background color is chosen, else the provided string is used as color value
	BrandButton        string `json:"brandButton,omitempty"`        // Decides if the pay button should have the same color as the brandColor. If set to false the hex color #01669F will be used as a default
	PayButtonTextColor string `json:"payButtonTextColor,omitempty"` // The color (hex) of the pay button
	LogoSrc            string `json:"logoSrc,omitempty"`            // An SVG image provided by the merchant. The image needs to be uploaded by using the Datatrans Web Administration Tool
	InitialView        string `json:"initialView,omitempty"`        // Enum: "list" "grid"	Wheter the payment page shows the payment method selection as list (default) or as a grid
	BrandTitle         bool   `json:"brandTitle,omitempty"`         // If set to false and no logo is used (see logoSrc), the payment page header will be empty
}

type Redirect struct {
	SuccessUrl string `json:"successUrl,omitempty"` // The URL where the customer gets redirected to if the transaction was successful.
	CancelUrl  string `json:"cancelUrl,omitempty"`  // The URL where the customer gets redirected to if the transaction was canceled.
	ErrorUrl   string `json:"errorUrl,omitempty"`   // The URL where the customer gets redirected to if an error occurred.
	// If the payment is started within an iframe or when using the Lightbox
	// Mode, use value _top. This ensures a proper browser flow for payment
	// methods who need a redirect.
	StartTarget string `json:"startTarget,omitempty"`
	// If the payment is started within an iframe or when using the Lightbox
	// Mode, use _top if the redirect URLs should be opened full screen when
	// payment returns from a 3rd party (for example 3D).
	ReturnTarget string `json:"returnTarget,omitempty"`
	// The preferred HTTP method for the redirect request (GET or POST). When
	// using GET as a method, the query string parameter datatransTrxId will be
	// added to the corresponding return url upon redirection. In case of POST,
	// all the query parameters from the corresponding return url will be moved
	// to the application/x-www-form-urlencoded body of the redirection request
	// along with the added datatransTrxId parameter.
	Method string `json:"method,omitempty"` // Default: "GET"	Enum: "GET" "POST"
}

type InitializeOption struct {
	// Whether an alias should be created for this transaction or not. If set to
	// true an alias will be created. This alias can then be used to initialize
	// or authorize a transaction. One possible use case is to charge the card of
	// an existing (registered) cardholder.
	CreateAlias            bool   `json:"createAlias"`
	ReturnMaskedCardNumber bool   `json:"returnMaskedCardNumber"` // Whether to return the masked card number. Format: 520000xxxxxx0080
	ReturnCustomerCountry  bool   `json:"returnCustomerCountry"`  // If set to true, the country of the customers issuer will be returned.
	AuthenticationOnly     bool   `json:"authenticationOnly"`     // Whether to only authenticate the transaction (3D process only). If set to true, the actual authorization will not take place.
	RememberMe             string `json:"rememberMe"`             // Enum: "true" "checked"	Whether to show a checkbox on the payment page to let the customer choose if they want to save their card information.
	ReturnMobileToken      bool   `json:"returnMobileToken"`      // Indicates that a mobile token should be created. This is needed when using our Mobile SDKs.
}

type RequestReconciliationsSale struct {
	Date          time.Time `json:"date"`
	TransactionID string    `json:"transactionId"`
	Currency      string    `json:"currency"`
	Amount        int       `json:"amount"`
	Type          string    `json:"type"`
	Refno         string    `json:"refno"`
}

type ResponseReconciliationsSale struct {
	TransactionID string    `json:"transactionId"`
	SaleDate      time.Time `json:"saleDate"`
	ReportedDate  time.Time `json:"reportedDate"`
	MatchResult   string    `json:"matchResult"`
}

type RequestReconciliationsSales struct {
	Sales []RequestReconciliationsSale `json:"sales"`
}

type ResponseReconciliationsSales struct {
	Sales []ResponseReconciliationsSale `json:"sales"`
}
