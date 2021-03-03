package datatrans

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s\n%#v", err, err)
	}
}

func Test_extractTimeAndHash(t *testing.T) {
	tests := []struct {
		name        string
		headerValue string
		wantTime    string
		wantS0hash  []byte
	}{
		{
			name:        "ok",
			headerValue: "t=1559303131511,s0=33819a1220fd8e38fc5bad3f57ef31095fac0deb38c001ba347e694f48ffe2fc",
			wantTime:    "1559303131511",
			wantS0hash:  []byte{0x33, 0x81, 0x9a, 0x12, 0x20, 0xfd, 0x8e, 0x38, 0xfc, 0x5b, 0xad, 0x3f, 0x57, 0xef, 0x31, 0x9, 0x5f, 0xac, 0xd, 0xeb, 0x38, 0xc0, 0x1, 0xba, 0x34, 0x7e, 0x69, 0x4f, 0x48, 0xff, 0xe2, 0xfc},
		},
		{
			name:        "empty vals",
			headerValue: "t=,s0=",
		},
		{
			name: "empty",
		},
		{
			name:        "missing comma",
			headerValue: "t=1559303131511s0=33",
		},
		{
			name:        "comma begin",
			headerValue: ",t=1559303131511s0=33",
		},
		{
			name:        "comma end",
			headerValue: "t=1559303131511s0=33,",
		},
		{
			name:        "comma only",
			headerValue: ",",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTime, gotS0hash := extractTimeAndHash(tt.headerValue)
			if gotTime != tt.wantTime {
				t.Errorf("extractTimeAndHash() gotTime = %v, want %v", gotTime, tt.wantTime)
			}
			if !bytes.Equal(gotS0hash, tt.wantS0hash) {
				t.Errorf("extractTimeAndHash() gotS0hash = %x, want %x", gotS0hash, tt.wantS0hash)
			}
		})
	}
}

func TestValidateWebhook(t *testing.T) {
	sign2Key := []byte(`asdfasd^%@^&%fa`)
	const timeStr = `1559303131511`

	mw, err := ValidateWebhook(WebhookOption{
		Sign2HMACKey: "617364666173645e25405e26256661",
	})
	must(t, err)

	const datatransBody = `{"transactionId": "210215103042148501"}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(datatransBody))

	ht := hmac.New(sha256.New, sign2Key)
	fmt.Fprintf(ht, "%s%s", timeStr, datatransBody)
	r.Header.Set("Datatrans-Signature", fmt.Sprintf("t=%s,s0=%x", timeStr, ht.Sum(nil)))

	w := httptest.NewRecorder()
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "success")
	})).ServeHTTP(w, r)

	if w.Body.String() != "success" {
		t.Error("something is wrong")
	}
}
