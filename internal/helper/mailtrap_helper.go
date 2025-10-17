package helper

import (
	"bytes"
	"encoding/json"
	"fiber-boiler-plate/config"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MailtrapEmailRequest struct {
	From struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"from"`
	To []struct {
		Email string `json:"email"`
	} `json:"to"`
	TemplateUUID      string                 `json:"template_uuid"`
	TemplateVariables map[string]interface{} `json:"template_variables"`
}

type MailtrapResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type MailtrapClient struct {
	apiToken  string
	accountID string
	domain    string
	client    *http.Client
	logger    *Logger
}

func NewMailtrapClient(cfg *config.MailtrapConfig) *MailtrapClient {
	return &MailtrapClient{
		apiToken:  cfg.APIToken,
		accountID: cfg.AccountID,
		domain:    cfg.Domain,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: NewLogger(),
	}
}

func (mc *MailtrapClient) SendOTPEmail(recipientEmail string, otpCode string, templateID string) error {
	// Validate inputs
	if recipientEmail == "" {
		return fmt.Errorf("recipient email cannot be empty")
	}
	if otpCode == "" {
		return fmt.Errorf("OTP code cannot be empty")
	}
	if templateID == "" {
		return fmt.Errorf("template ID cannot be empty")
	}

	url := "https://send.api.mailtrap.io/api/send"

	req := MailtrapEmailRequest{}
	req.From.Email = "noreply@builtwithafi.web.id"
	req.From.Name = "Invento"
	req.To = []struct {
		Email string `json:"email"`
	}{
		{Email: recipientEmail},
	}
	req.TemplateUUID = templateID
	req.TemplateVariables = map[string]interface{}{
		"otp_code": otpCode,
		"otp":      otpCode,
		"code":     otpCode,
		"otpCode":  otpCode,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mc.apiToken))

	resp, err := mc.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("mailtrap error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var mailResp MailtrapResponse
	if err := json.Unmarshal(respBody, &mailResp); err != nil {
		return err
	}

	if !mailResp.Success {
		return fmt.Errorf("mailtrap error: %v", mailResp.Data)
	}

	return nil
}

type ResendRateLimiter struct {
	maxResends      int
	cooldownSeconds int
}

func NewResendRateLimiter(maxResends int, cooldownSeconds int) *ResendRateLimiter {
	return &ResendRateLimiter{
		maxResends:      maxResends,
		cooldownSeconds: cooldownSeconds,
	}
}

func (rrl *ResendRateLimiter) CanResend(resendCount int, lastResendAt *time.Time) (bool, int64) {
	if resendCount >= rrl.maxResends {
		return false, 0
	}

	if lastResendAt == nil {
		return true, 0
	}

	elapsed := time.Since(*lastResendAt).Seconds()
	waitTime := int64(rrl.cooldownSeconds) - int64(elapsed)

	if waitTime > 0 {
		return false, waitTime
	}

	return true, 0
}

func (rrl *ResendRateLimiter) GetMaxResends() int {
	return rrl.maxResends
}

func (rrl *ResendRateLimiter) GetCooldownSeconds() int {
	return rrl.cooldownSeconds
}
