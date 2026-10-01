package strings_utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"unicode"
)

// Runes are good for manipulating string

func IsValidLength(v string, min, max int) bool {
	s := strings.TrimSpace(v)
	runes := []rune(s)
	length := len(runes)
	if length < min || length > max {
		return false
	}
	return true
}

func Capitalize(s string) string {
	words := strings.Fields(s)

	for i, word := range words {
		runes := []rune(word)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		words[i] = string(runes)
	}

	return strings.Join(words, " ")
}

func Substring(s string, max int) string {
	v := strings.TrimSpace(s)
	runes := []rune(v)
	length := len(runes)
	if max <= 0 {
		return ""
	}
	if length <= max {
		return v
	}
	return string(runes[:max])
}

func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	email = strings.ToLower(email)

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	return addr.Address == email
}

func ObfuscateEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	local := []rune(parts[0])
	domain := parts[1]

	if len(local) <= 2 {
		return strings.Repeat("*", len(local)) + "@" + domain
	}

	return string(local[0]) +
		strings.Repeat("*", len(local)-2) +
		string(local[len(local)-1]) +
		"@" + domain
}

func NormalizeWebsiteURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return "", errors.New("website is required")
	}

	if !strings.HasPrefix(raw, "http://") &&
		!strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("invalid website URL")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("only http and https URLs are allowed")
	}

	if u.Hostname() == "" {
		return "", errors.New("missing hostname")
	}

	return u.String(), nil
}

func FormatNullString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func HashToken(token string) string {
	secret := []byte(os.Getenv("SECRET_KEY"))
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}
