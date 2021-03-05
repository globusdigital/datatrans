package datatrans

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// https://docs.datatrans.ch/docs/api-endpoints
const (
	endpointURLSandBox    = `https://api.sandbox.datatrans.com`
	endpointURLProduction = `https://api.datatrans.com`

	pathBase                     = "/v1/transactions"
	pathStatus                   = pathBase + "/%s"
	pathCredit                   = pathBase + "/%s/credit"
	pathCreditAuthorize          = pathBase + "/credit"
	pathCancel                   = pathBase + "/%s/cancel"
	pathSettle                   = pathBase + "/%s/settle"
	pathValidate                 = pathBase + "/validate"
	pathAuthorizeTransaction     = pathBase + "/%s/authorize"
	pathAuthorize                = pathBase + "/authorize"
	pathInitialize               = pathBase
	pathSecureFields             = pathBase + "/secureFields"
	pathSecureFieldsUpdate       = pathBase + "/secureFields/%s"
	pathAliases                  = pathBase + "/aliases"
	pathAliasesDelete            = pathBase + "/aliases/%s"
	pathReconciliationsSales     = "/v1/reconciliations/sales"
	pathReconciliationsSalesBulk = "/v1/reconciliations/sales/bulk"
)

type OptionMerchant struct {
	InternalID       string
	EnableProduction bool
	// https://docs.datatrans.ch/docs/api-endpoints#section-idempotency
	// If your request failed to reach our servers, no idempotent result is saved
	// because no API endpoint processed your request. In such cases, you can
	// simply retry your operation safely. Idempotency keys remain stored for 3
	// minutes. After 3 minutes have passed, sending the same request together
	// with the previous idempotency key will create a new operation.
	EnableIdempotency  bool
	DisableRawJSONBody bool
	MerchantID         string // basic auth user
	Password           string // basic auth pw
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
	doFn              OptionHTTPRequestFn
	merchants         map[string]OptionMerchant // string = your custom merchant ID
	currentInternalID string
	internalIDFound   bool
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
	if c.doFn == nil {
		c.doFn = (&http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		}).Do
	}
	// see if we have a default one, otherwise you always have to call WithMerchant.
	_, c.internalIDFound = c.merchants[""]
	return c, nil
}

// WithMerchant sets an ID and returns a shallow clone of the client.
func (c *Client) WithMerchant(internalID string) *Client {
	c2 := *c
	c2.currentInternalID = internalID
	_, c2.internalIDFound = c2.merchants[internalID]
	return &c2
}

func (c *Client) do(req *http.Request, v interface{}) error {
	internalID := c.currentInternalID
	if !c.internalIDFound {
		return fmt.Errorf("ClientID %q not found in list of merchants", internalID)
	}

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
	if set, ok := v.(rawJSONBodySetter); !c.merchants[internalID].DisableRawJSONBody && ok {
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

func (c *Client) prepareJSONReq(ctx context.Context, method, path string, postData interface{}) (*http.Request, error) {
	internalID := c.currentInternalID

	var r io.Reader
	var jsonBytes []byte
	if postData != nil {
		var err error
		jsonBytes, err = MarshalJSON(postData)
		if err != nil {
			return nil, fmt.Errorf("ClientID:%q: failed to json marshal HTTP request: %w", internalID, err)
		}
		r = bytes.NewReader(jsonBytes)
	}
	host := endpointURLSandBox
	if c.merchants[internalID].EnableProduction {
		host = endpointURLProduction
	}

	req, err := http.NewRequestWithContext(ctx, method, host+path, r)
	if err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to create HTTP request: %w", internalID, err)
	}
	if postData != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if method == http.MethodPost && c.merchants[internalID].EnableIdempotency {
		// not quite happy with this
		// https://docs.datatrans.ch/docs/api-endpoints#section-idempotency
		fh := fnv.New64a()
		_, _ = fh.Write([]byte(internalID + host + path))
		_, _ = fh.Write(jsonBytes)
		req.Header.Set("Idempotency-Key", hex.EncodeToString(fh.Sum(nil)))
	}

	return req, nil
}

// Status allows once a transactionId has been received the status can be checked
// with the Status API.
func (c *Client) Status(ctx context.Context, transactionID string) (*ResponseStatus, error) {
	if transactionID == "" {
		return nil, fmt.Errorf("transactionID cannot be empty")
	}
	internalID := c.currentInternalID
	host := endpointURLSandBox
	if c.merchants[internalID].EnableProduction {
		host = endpointURLProduction
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(host+pathStatus, transactionID), nil)
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
func (c *Client) Credit(ctx context.Context, transactionID string, rc RequestCredit) (*ResponseCardMasked, error) {
	if transactionID == "" || rc.Currency == "" || rc.RefNo == "" {
		return nil, fmt.Errorf("neither currency nor refno nor transactionID can be empty")
	}

	req, err := c.prepareJSONReq(ctx, http.MethodPost, fmt.Sprintf(pathCredit, transactionID), rc)
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
func (c *Client) CreditAuthorize(ctx context.Context, rca RequestCreditAuthorize) (*ResponseCardMasked, error) {
	if rca.Currency == "" || rca.RefNo == "" || rca.Amount == 0 {
		return nil, fmt.Errorf("neither currency nor refno nor amount can be empty")
	}

	req, err := c.prepareJSONReq(ctx, http.MethodPost, pathCreditAuthorize, rca)
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
func (c *Client) Cancel(ctx context.Context, transactionID string, refno string) error {
	if transactionID == "" || refno == "" {
		return fmt.Errorf("neither transactionID nor refno can be empty")
	}
	req, err := c.prepareJSONReq(ctx, http.MethodPost, fmt.Sprintf(pathCancel, transactionID), struct {
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
func (c *Client) Settle(ctx context.Context, transactionID string, rs RequestSettle) error {
	if transactionID == "" || rs.Amount == 0 || rs.Currency == "" || rs.RefNo == "" {
		return fmt.Errorf("neither transactionID nor refno nor amount nor currency can be empty")
	}
	req, err := c.prepareJSONReq(ctx, http.MethodPost, fmt.Sprintf(pathSettle, transactionID), rs)
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
func (c *Client) ValidateAlias(ctx context.Context, rva RequestValidateAlias) (*ResponseCardMasked, error) {
	if rva.Currency == "" || rva.RefNo == "" {
		return nil, fmt.Errorf("neither currency nor refno can be empty")
	}
	req, err := c.prepareJSONReq(ctx, http.MethodPost, pathValidate, rva)
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
func (c *Client) AuthorizeTransaction(ctx context.Context, transactionID string, rva RequestAuthorizeTransaction) (*ResponseAuthorize, error) {
	if transactionID == "" || rva.RefNo == "" {
		return nil, fmt.Errorf("neither transactionID nor refno can be empty")
	}
	req, err := c.prepareJSONReq(ctx, http.MethodPost, fmt.Sprintf(pathAuthorizeTransaction, transactionID), rva)
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
func (c *Client) Authorize(ctx context.Context, rva RequestAuthorize) (*ResponseCardMasked, error) {
	if rva.Amount == 0 || rva.Currency == "" || rva.RefNo == "" {
		return nil, fmt.Errorf("neither transactionID nor amount nor currency nor refno can be empty")
	}
	req, err := c.prepareJSONReq(ctx, http.MethodPost, pathAuthorize, rva)
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
func (c *Client) Initialize(ctx context.Context, rva RequestInitialize) (*ResponseInitialize, error) {
	if rva.Amount == 0 || rva.Currency == "" || rva.RefNo == "" {
		return nil, fmt.Errorf("neither amount nor currency nor refno can be empty")
	}
	req, err := c.prepareJSONReq(ctx, http.MethodPost, pathInitialize, rva)
	if err != nil {
		return nil, err
	}

	var ri ResponseInitialize
	if err := c.do(req, &ri); err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return &ri, nil
}

// InitializeSecureFields initializes a Secure Fields transaction. Proceed with
// the steps below to process Secure Fields payment transactions.
// https://api-reference.datatrans.ch/#operation/secureFieldsInit
func (c *Client) SecureFieldsInit(ctx context.Context, rva RequestSecureFieldsInit) (*ResponseInitialize, error) {
	if rva.Amount == 0 || rva.Currency == "" || rva.ReturnUrl == "" {
		return nil, fmt.Errorf("neither amount nor currency nor returnURL can be empty")
	}
	req, err := c.prepareJSONReq(ctx, http.MethodPost, pathSecureFields, rva)
	if err != nil {
		return nil, err
	}

	var ri ResponseInitialize
	if err := c.do(req, &ri); err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return &ri, nil
}

// SecureFieldsUpdate use this API to update the amount of a Secure Fields
// transaction. This action is only allowed before the 3D process. At least one
// property must be updated.
// https://api-reference.datatrans.ch/#operation/secure-fields-update
func (c *Client) SecureFieldsUpdate(ctx context.Context, transactionID string, rva RequestSecureFieldsUpdate) error {
	if rva.Amount == 0 || rva.Currency == "" {
		return fmt.Errorf("neither amount nor currency nor returnURL can be empty")
	}
	req, err := c.prepareJSONReq(ctx, http.MethodPatch, fmt.Sprintf(pathSecureFieldsUpdate, transactionID), rva)
	if err != nil {
		return err
	}

	if err := c.do(req, nil); err != nil {
		return fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return nil
}

// AliasConvert converts a legacy (numeric or masked) alias to the most recent
// alias format.
func (c *Client) AliasConvert(ctx context.Context, legacyAlias string) (string, error) {
	if legacyAlias == "" {
		return "", fmt.Errorf("legacyAlias cannot be empty")
	}
	req, err := c.prepareJSONReq(ctx, http.MethodPost, pathAliases, struct {
		LegacyAlias string `json:"legacyAlias"`
	}{
		LegacyAlias: legacyAlias,
	})
	if err != nil {
		return "", err
	}
	var resp struct {
		Alias string `json:"alias"`
	}
	if err := c.do(req, &resp); err != nil {
		return "", fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return resp.Alias, nil
}

// AliasDelete deletes an alias with immediate effect. The alias will no longer
// be recognized if used later with any API call.
func (c *Client) AliasDelete(ctx context.Context, alias string) error {
	if alias == "" {
		return fmt.Errorf("alias cannot be empty")
	}
	req, err := c.prepareJSONReq(ctx, http.MethodDelete, fmt.Sprintf(pathAliasesDelete, alias), nil)
	if err != nil {
		return err
	}
	if err := c.do(req, nil); err != nil {
		return fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return nil
}

// ReconciliationsSales reports a sale. When using reconciliation, use this API
// to report a sale. The matching is based on the transactionId.
func (c *Client) ReconciliationsSales(ctx context.Context, sale RequestReconciliationsSale) (*ResponseReconciliationsSale, error) {
	req, err := c.prepareJSONReq(ctx, http.MethodPost, pathReconciliationsSales, sale)
	if err != nil {
		return nil, err
	}
	var rrs ResponseReconciliationsSale
	if err := c.do(req, &rrs); err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return &rrs, nil
}

// ReconciliationsSalesBulk reports bulk sales. When using reconciliation, use
// this API to report multiples sales with a single API call. The matching is
// based on the transactionId.
func (c *Client) ReconciliationsSalesBulk(ctx context.Context, sales RequestReconciliationsSales) (*ResponseReconciliationsSales, error) {
	req, err := c.prepareJSONReq(ctx, http.MethodPost, pathReconciliationsSalesBulk, sales)
	if err != nil {
		return nil, err
	}
	var rrs ResponseReconciliationsSales
	if err := c.do(req, &rrs); err != nil {
		return nil, fmt.Errorf("ClientID:%q: failed to execute HTTP request: %w", c.currentInternalID, err)
	}
	return &rrs, nil
}
