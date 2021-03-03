package datatrans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	pathBase                 = "/v1/transactions"
	pathStatus               = pathBase + "/%s"
	pathCredit               = pathBase + "/%s/credit"
	pathCreditAuthorize      = pathBase + "/credit"
	pathCancel               = pathBase + "/%s/cancel"
	pathSettle               = pathBase + "/%s/settle"
	pathValidate             = pathBase + "/validate"
	pathAuthorizeTransaction = pathBase + "/%s/authorize"
	pathAuthorize            = pathBase + "/authorize"
	pathInitialize           = pathBase
)

type OptionMerchant struct {
	InternalID string
	Server     string
	MerchantID string // basic auth user
	Password   string // basic auth pw
}

func (m OptionMerchant) apply(c *Client) error {
	if _, ok := c.merchants[m.InternalID]; ok {
		return fmt.Errorf("InternalID %q already exists", m.InternalID)
	}
	c.merchants[m.InternalID] = m
	return nil
}

type OptionHTTPRequestFn func(req *http.Request) (*http.Response, error)

func (fn OptionHTTPRequestFn) apply(c *Client) error {
	c.doFn = fn
	return nil
}

type Client struct {
	copyRawResponseBody bool
	doFn                OptionHTTPRequestFn
	merchants           map[string]OptionMerchant // string = your custom merchant ID
	currentInternalID   string
}

type Option interface {
	apply(*Client) error
}

func MakeClient(opts ...Option) (Client, error) {
	c := Client{
		merchants: make(map[string]OptionMerchant, 3),
	}
	for _, opt := range opts {
		if err := opt.apply(&c); err != nil {
			return Client{}, err
		}
	}
	if len(c.merchants) == 0 {
		return Client{}, fmt.Errorf("no merchants applied")
	}

	return c, nil
}

// WithMerchant sets an ID and returns a shallow clone of the client.
func (c *Client) WithMerchant(internalID string) *Client {
	c2 := *c
	c2.currentInternalID = internalID
	return &c2
}

func (c *Client) do(req *http.Request, v interface{}) error {
	internalID := c.currentInternalID
	req.SetBasicAuth(c.merchants[internalID].MerchantID, c.merchants[internalID].Password)
	resp, err := c.doFn(req)
	defer closeResponse(resp)
	if err != nil {
		return fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", internalID, err)
	}

	var buf bytes.Buffer
	body := io.TeeReader(resp.Body, &buf)
	dec := json.NewDecoder(body)

	if !c.isSuccess(resp.StatusCode) {
		var errResp ErrorResponse
		if err := dec.Decode(&errResp); err != nil {
			return fmt.Errorf("ClientID:%q: failed to unmarshal HTTP error response: %w", internalID, err)
		}
		errResp.HTTPStatusCode = resp.StatusCode
		return errResp
	}
	if v != nil {
		if err := dec.Decode(v); err != nil {
			return fmt.Errorf("ClientID:%q: failed to unmarshal HTTP error response: %w", internalID, err)
		}
	}
	if ri, ok := v.(*ResponseInitialize); ok {
		if loc := resp.Header.Get("Location"); loc != "" {
			ri.Location = loc
		}
	}
	if set, ok := v.(rawJSONBodySetter); ok {
		set.setJSONRawBody(buf.Bytes())
	}

	return nil
}

func (c *Client) isSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

func closeResponse(r *http.Response) {
	if r == nil || r.Body == nil {
		return
	}
	_, _ = io.Copy(ioutil.Discard, r.Body)
	_ = r.Body.Close()
}

// MarshalJSON encodes the postData struct to json but also can merge custom
// settings into the final JSON. This function is called before sending the
// request to datatrans. Function exported for debug reasons.
func MarshalJSON(postData interface{}) ([]byte, error) {
	jsonBytes, err := json.Marshal(postData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal postData: %w", err)
	}

	// this steps merges two different Go types into one JS object.
	if cfg, ok := postData.(customFieldsGetter); ok {
		custFields := cfg.getCustomFields()
		if len(custFields) == 0 {
			return jsonBytes, nil
		}

		postDataMap := map[string]interface{}{}
		if err := json.Unmarshal(jsonBytes, &postDataMap); err != nil {
			return nil, fmt.Errorf("failed to Unmarshal postData raw bytes: %w", err)
		}
		for k, v := range custFields {
			postDataMap[k] = v // overwrites existing data from postData struct
		}
		jsonBytes, err = json.Marshal(postDataMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal postDataMap: %w", err)
		}
	}

	return jsonBytes, nil
}

func (c *Client) preparePostJSONReq(path string, postData interface{}) (*http.Request, error) {
	internalID := c.currentInternalID

	jsonBytes, err := MarshalJSON(postData)
	if err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to json marshal HTTP request: %w", internalID, err)
	}

	req, err := http.NewRequest(http.MethodPost, c.merchants[internalID].Server+path, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to create HTTP request: %w", internalID, err)
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// Status allows once a transactionId has been received the status can be checked
// with the Status API.
func (c *Client) Status(transactionID string) (*ResponseStatus, error) {
	if transactionID == "" {
		return nil, fmt.Errorf("transactionID cannot be empty")
	}
	internalID := c.currentInternalID
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(c.merchants[internalID].Server+pathStatus, transactionID), nil)
	if err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to create HTTP request: %w", internalID, err)
	}

	var respStatus ResponseStatus
	if err := c.do(req, &respStatus); err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", internalID, err)
	}

	return &respStatus, nil
}

// Credit uses the credit API to credit a transaction which is in status settled.
// The previously settled amount must not be exceeded.
func (c *Client) Credit(transactionID string, rc RequestCredit) (*ResponseCardMasked, error) {
	if transactionID == "" || rc.Currency == "" || rc.RefNo == "" {
		return nil, fmt.Errorf("neither currency nor refno nor transactionID can be empty")
	}

	req, err := c.preparePostJSONReq(fmt.Sprintf(pathCredit, transactionID), rc)
	if err != nil {
		return nil, err
	}

	var respRefund ResponseCardMasked
	if err := c.do(req, &respRefund); err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}

	return &respRefund, nil
}

// CreditAuthorize allows to use this API to make a credit without referring to a
// previous authorization. This can be useful if you want to credit a cardholder
// when there was no debit.
func (c *Client) CreditAuthorize(rca RequestCreditAuthorize) (*ResponseCardMasked, error) {
	if rca.Currency == "" || rca.RefNo == "" || rca.Amount == 0 {
		return nil, fmt.Errorf("neither currency nor refno nor amount can be empty")
	}

	req, err := c.preparePostJSONReq(pathCreditAuthorize, rca)
	if err != nil {
		return nil, err
	}

	var respRefund ResponseCardMasked
	if err := c.do(req, &respRefund); err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return &respRefund, nil
}

// Cancel API can be used to release the blocked amount from an authorization.
// The transaction must either be in status authorized or settled. The
// transactionId is needed to cancel an authorization.
// https://api-reference.datatrans.ch/#operation/cancel
func (c *Client) Cancel(transactionID string, refno string) error {
	if transactionID == "" || refno == "" {
		return fmt.Errorf("neither transactionID nor refno can be empty")
	}
	req, err := c.preparePostJSONReq(fmt.Sprintf(pathCancel, transactionID), struct {
		Refno string `json:"refno"`
	}{
		Refno: refno,
	})
	if err != nil {
		return err
	}

	if err := c.do(req, nil); err != nil {
		return fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return nil
}

// Settle request is often also referred to as “Capture” or “Clearing”. It can be
// used for the settlement of previously authorized transactions. The
// transactionId is needed to settle an authorization. Note: This API call is not
// needed if "autoSettle": true was used when initializing a transaction.
// https://api-reference.datatrans.ch/#operation/settle
func (c *Client) Settle(transactionID string, rs RequestSettle) error {
	if transactionID == "" || rs.Amount == 0 || rs.Currency == "" || rs.RefNo == "" {
		return fmt.Errorf("neither transactionID nor refno nor amount nor currency can be empty")
	}
	req, err := c.preparePostJSONReq(fmt.Sprintf(pathSettle, transactionID), rs)
	if err != nil {
		return err
	}

	if err := c.do(req, nil); err != nil {
		return fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return nil
}

// ValidateAlias an existing alias can be validated at any time with the
// transaction validate API. No amount will be blocked on the customers account.
// Only credit cards (including Apple Pay and Google Pay), PFC, KLN and PAP
// support validation of an existing alias.
// https://api-reference.datatrans.ch/#operation/validate
func (c *Client) ValidateAlias(rva RequestValidateAlias) (*ResponseCardMasked, error) {
	if rva.Currency == "" || rva.RefNo == "" {
		return nil, fmt.Errorf("neither currency nor refno can be empty")
	}
	req, err := c.preparePostJSONReq(pathValidate, rva)
	if err != nil {
		return nil, err
	}

	var rcm ResponseCardMasked
	if err := c.do(req, &rcm); err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return &rcm, nil
}

// AuthorizeTransaction an authenticated transaction. If during the initialization of a
// transaction the parameter option.authenticationOnly was set to true, this API
// can be used to authorize an already authenticated (3D) transaction.
// https://api-reference.datatrans.ch/#operation/authorize-split
func (c *Client) AuthorizeTransaction(transactionID string, rva RequestAuthorizeTransaction) (*ResponseAuthorize, error) {
	if transactionID == "" || rva.RefNo == "" {
		return nil, fmt.Errorf("neither transactionID nor refno can be empty")
	}
	req, err := c.preparePostJSONReq(fmt.Sprintf(pathAuthorizeTransaction, transactionID), rva)
	if err != nil {
		return nil, err
	}

	var rcm ResponseAuthorize
	if err := c.do(req, &rcm); err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return &rcm, nil
}

// Authorize a transaction. Use this API to make an authorization without user
// interaction. (For example merchant initiated transactions with an alias)
// Depending on the payment method, different parameters are mandatory. Refer to
// the payment method specific objects (for example PAP) to see which parameters
// so send. For credit cards, the card object can be used.
// https://api-reference.datatrans.ch/#operation/authorize
func (c *Client) Authorize(rva RequestAuthorize) (*ResponseCardMasked, error) {
	if rva.Amount == 0 || rva.Currency == "" || rva.RefNo == "" {
		return nil, fmt.Errorf("neither transactionID nor amount nor currency nor refno can be empty")
	}
	req, err := c.preparePostJSONReq(pathAuthorize, rva)
	if err != nil {
		return nil, err
	}

	var rcm ResponseCardMasked
	if err := c.do(req, &rcm); err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return &rcm, nil
}

// Initialize a transaction. Securely send all the needed parameters to the
// transaction initialization API. The result of this API call is a HTTP 201
// status code with a transactionId in the response body and the Location header
// set. If you want to use the payment page redirect mode to collect the payment
// details, the browser needs to be redirected to this URL to continue with the
// transaction. Following the link provided in the Location header will raise the
// Datatrans Payment Page with all the payment methods available for the given
// merchantId. If you want to limit the number of payment methods, the
// paymentMethod array can be used.
func (c *Client) Initialize(rva RequestInitialize) (*ResponseInitialize, error) {
	if rva.Amount == 0 || rva.Currency == "" || rva.RefNo == "" {
		return nil, fmt.Errorf("neither transactionID nor amount nor currency nor refno can be empty")
	}
	req, err := c.preparePostJSONReq(pathInitialize, rva)
	if err != nil {
		return nil, err
	}

	var ri ResponseInitialize
	if err := c.do(req, &ri); err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return &ri, nil
}
