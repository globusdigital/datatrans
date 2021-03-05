package datatrans_test

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/globusdigital/datatrans"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s\n%#v", err, err)
	}
}

func mockResponse(
	t *testing.T,
	status int,
	body string,
	testReq func(t *testing.T, req *http.Request),
) func(req *http.Request) (*http.Response, error) {
	rc := ioutil.NopCloser(strings.NewReader(body))
	if strings.HasSuffix(body, ".json") {
		fp, err := os.Open(body)
		if err != nil {
			t.Fatal(err)
		}
		rc = fp
	}
	if testReq == nil {
		testReq = func(t *testing.T, req *http.Request) {}
	}

	return func(req *http.Request) (*http.Response, error) {
		testReq(t, req)
		return &http.Response{
			StatusCode: status,
			Body:       rc,
		}, nil
	}
}

func TestClient_Status(t *testing.T) {
	c, err := datatrans.MakeClient(
		datatrans.OptionHTTPRequestFn(mockResponse(t, 200, "testdata/status_response.json", nil)),
		datatrans.OptionMerchant{
			MerchantID: "322342",
			Password:   "32168",
		},
	)
	must(t, err)

	rs, err := c.Status(context.Background(), "3423423423")
	must(t, err)
	if rs.TransactionID != "210215103042148501" {
		t.Errorf("incorrect TransactionID:%q", rs.TransactionID)
	}
}

func TestClient_Initialize(t *testing.T) {
	c, err := datatrans.MakeClient(
		datatrans.OptionHTTPRequestFn(mockResponse(t, 200, `{"transactionId": "210215103033478409"}`, func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodPost {
				t.Error("not a post request")
			}
			if req.Header.Get("Content-Type") != "application/json" {
				t.Error("invalid content type")
			}
			if k := req.Header.Get("Idempotency-Key"); k != "c0476553a7e7da70" {
				t.Errorf("invalid Idempotency-Key: %q", k)
			}

			u, p, _ := req.BasicAuth()
			if u != "322342" {
				t.Error("invalid basic username")
			}
			if p != "sfdgsdfg" {
				t.Error("invalid basic password")
			}
			var buf bytes.Buffer
			buf.ReadFrom(req.Body)

			const wantBody = `{"alp":true,"amount":1337,"currency":"CHF","paymentMethods":["VIS","PFC"],"redirect":{"cancelUrl":"https://.../cancelPage.jsp","errorUrl":"https://.../errorPage.jsp","successUrl":"https://.../successPage.jsp"},"refno":"872732"}`
			if buf.String() != wantBody {
				t.Errorf("invalid body: %q", buf.String())
			}
		})),
		datatrans.OptionMerchant{
			EnableIdempotency: true,
			MerchantID:        "322342",
			Password:          "sfdgsdfg",
		},
	)
	must(t, err)

	ri := datatrans.RequestInitialize{
		Currency:       "CHF",
		RefNo:          "872732",
		Amount:         1337,
		Language:       "",
		PaymentMethods: []string{"VIS", "PFC"},
		Redirect: &datatrans.Redirect{
			SuccessUrl: "https://.../successPage.jsp",
			CancelUrl:  "https://.../cancelPage.jsp",
			ErrorUrl:   "https://.../errorPage.jsp",
		},
		CustomFields: map[string]interface{}{
			"alp": true,
		},
	}
	t.Run("successful request", func(t *testing.T) {
		rs, err := c.Initialize(context.Background(), ri)
		must(t, err)

		want := &datatrans.ResponseInitialize{TransactionId: "210215103033478409", RawJSONBody: datatrans.RawJSONBody{0x7b, 0x22, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x22, 0x3a, 0x20, 0x22, 0x32, 0x31, 0x30, 0x32, 0x31, 0x35, 0x31, 0x30, 0x33, 0x30, 0x33, 0x33, 0x34, 0x37, 0x38, 0x34, 0x30, 0x39, 0x22, 0x7d}}
		if !reflect.DeepEqual(rs, want) {
			t.Error("invalid response")
		}
	})

	t.Run("invalid merchant", func(t *testing.T) {
		rs, err := c.WithMerchant("sadfasdf").Initialize(context.Background(), ri)
		if rs != nil {
			t.Error("response should be nil")
		}
		if err == nil {
			t.Fatal("expected an error")
		}
		t.Log(err)
	})
}

func TestMarshalJSON(t *testing.T) {
	ri := datatrans.RequestInitialize{
		Currency:   "CHF",
		RefNo:      "234234",
		RefNo2:     "",
		AutoSettle: true,
		Amount:     123,
		Language:   "DE",
		CustomFields: map[string]interface{}{
			"twi": map[string]interface{}{
				"alias": "ZGZhc2RmYXNkZmFzZGZhc2Q=",
			},
		},
	}
	data, err := datatrans.MarshalJSON(ri)
	must(t, err)
	const wantJSON = `{"amount":123,"autoSettle":true,"currency":"CHF","language":"DE","refno":"234234","twi":{"alias":"ZGZhc2RmYXNkZmFzZGZhc2Q="}}`
	if string(data) != wantJSON {
		t.Errorf("\nWant: %s\nHave: %s", wantJSON, data)
	}
}

func TestClient_AliasDelete_Error(t *testing.T) {
	c, err := datatrans.MakeClient(
		datatrans.OptionHTTPRequestFn(mockResponse(t, 400, `{"error": {"code": "ALIAS_NOT_FOUND"}}`, func(t *testing.T, req *http.Request) {
			if req.Method != http.MethodDelete {
				t.Error("not a delete request")
			}
			if req.Header.Get("Content-Type") == "application/json" {
				t.Error("invalid content type")
			}

			u, p, _ := req.BasicAuth()
			if u != "322342" {
				t.Error("invalid basic username")
			}
			if p != "sfdgsdfg" {
				t.Error("invalid basic password")
			}
		})),
		datatrans.OptionMerchant{
			MerchantID: "322342",
			Password:   "sfdgsdfg",
		},
	)
	must(t, err)

	err = c.AliasDelete(context.Background(), "3469efdbbdcb043e56b19ffca69a8be0c5524d89")

	var detailErr datatrans.ErrorResponse
	errors.As(err, &detailErr)
	if !reflect.DeepEqual(detailErr, datatrans.ErrorResponse{HTTPStatusCode: 400, ErrorDetail: datatrans.ErrorDetail{Code: "ALIAS_NOT_FOUND", Message: ""}}) {
		t.Error("errors not equal")
	}
}
