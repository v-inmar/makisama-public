package captcha_utils

import (
	"app/internal/constants"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func CheckCaptcha(token string) error {
	secret := os.Getenv("TURNSTILE_SECRET")
	resp, err := http.PostForm(
		constants.TurnstileCaptchaSiteVerify,
		url.Values{
			"secret":   {secret},
			"response": {token},
		},
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var result struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("captcha failed: %v", result.ErrorCodes)
	}

	return nil
}
