package datatrans

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%#v", err)
	}
}

func mockResponse(status int, body string) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Body:       ioutil.NopCloser(strings.NewReader(body)),
		}, nil
	}
}

func TestClient_Status(t *testing.T) {
	_,_ = MakeClient(
		OptionHTTPRequestFn((&http.Client{
			Timeout: 30*time.Second,
		}).Do),
		OptionMerchant{
			Server:     "https://api.sandbox.datatrans.com",
			MerchantID: "32234323242",
			Password:   "dbce0e6cfc012e475c843c1bbb0ca439a048fe8e",
		},
	)

	c, err := MakeClient(
		OptionHTTPRequestFn(mockResponse(200, "{}")),
		OptionMerchant{
			Server:     "http://localhost",
			MerchantID: "322342",
			Password:   "32168",
		},
	)
	must(t, err)



	rs, err := c.Status("3423423423")
	must(t, err)
	// TODO continue here
	t.Logf("%#v", rs)
}
