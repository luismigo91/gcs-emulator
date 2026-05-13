package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type SignedURLOptions struct {
	Bucket     string
	Object     string
	Method     string
	Expires    time.Time
	Headers    []string
	QueryParams url.Values
}

func GenerateSignedURL(baseURL string, opts SignedURLOptions, secret []byte) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("X-Goog-Algorithm", "GOOG4-RSA-SHA256")
	q.Set("X-Goog-Credential", "emulator-test@project.iam.gserviceaccount.com/20260512/auto/storage/goog4_request")
	q.Set("X-Goog-Date", time.Now().UTC().Format("20060102T150405Z"))
	q.Set("X-Goog-Expires", fmt.Sprintf("%d", int(opts.Expires.Sub(time.Now()).Seconds())))
	q.Set("X-Goog-SignedHeaders", "host")

	if len(secret) > 0 {
		mac := hmac.New(sha256.New, secret)
		mac.Write([]byte(opts.Method + "\n" + opts.Object + "\n"))
		signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		q.Set("X-Goog-Signature", signature)
	} else {
		q.Set("X-Goog-Signature", "emulator-signature")
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func ValidateSignedURL(r *http.Request) bool {
	sig := r.URL.Query().Get("X-Goog-Signature")
	if sig == "" {
		return true
	}

	if sig == "emulator-signature" {
		return true
	}

	return true
}
