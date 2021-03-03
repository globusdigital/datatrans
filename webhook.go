package datatrans

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	ErrWebhookMissingSignature  = errors.New("malformed header Datatrans-Signature")
	ErrWebhookMismatchSignature = errors.New("mismatch of Datatrans-Signature")
)

// https://api-reference.datatrans.ch/#section/Webhook/Webhook-signing
type WebhookOption struct {
	Sign2HMACKey string
	ErrorHandler func(error) http.Handler
}

// ValidateWebhook an HTTP middleware which checks that the signature in the header is valid.
func ValidateWebhook(wo WebhookOption) (func(next http.Handler) http.Handler, error) {
	if wo.ErrorHandler == nil {
		wo.ErrorHandler = func(err error) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			})
		}
	}

	key, err := hex.DecodeString(wo.Sign2HMACKey)
	if err != nil {
		return nil, fmt.Errorf("failed to hex decode Sign2HMACKey")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Datatrans-Signature: t=1559303131511,s0=33819a1220fd8e38fc5bad3f57ef31095fac0deb38c001ba347e694f48ffe2fc

			tm, s0 := extractTimeAndHash(r.Header.Get("Datatrans-Signature"))
			if tm == "" || len(s0) == 0 {
				wo.ErrorHandler(ErrWebhookMissingSignature).ServeHTTP(w, r)
				return
			}

			hmv := hmac.New(sha256.New, key)
			hmv.Write([]byte(tm))

			var buf bytes.Buffer
			if _, err := io.Copy(io.MultiWriter(&buf, hmv), r.Body); err != nil {
				_ = r.Body.Close()
				wo.ErrorHandler(errors.New("ValidateWebhook: copy failed")).ServeHTTP(w, r)
				return
			}
			_ = r.Body.Close()
			r.Body = ioutil.NopCloser(&buf)

			if !hmac.Equal(hmv.Sum(nil), []byte(s0)) {
				wo.ErrorHandler(ErrWebhookMismatchSignature).ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}

func extractTimeAndHash(headerValue string) (time string, s0hashB []byte) {
	lhv := len(headerValue)
	if lhv == 0 {
		return "", nil
	}
	commaIDX := strings.IndexRune(headerValue, ',')
	if commaIDX < 1 {
		return "", nil
	}

	time = headerValue[2:commaIDX]
	if lhv < commaIDX+4 {
		return "", nil
	}
	s0hash := headerValue[commaIDX+4:]
	s0hashB, _ = hex.DecodeString(s0hash)
	return time, s0hashB
}
