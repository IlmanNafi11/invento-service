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

// ResendClient handles email sending via Resend API
type ResendClient struct {
	apiKey    string
	fromEmail string
	fromName  string
	client    *http.Client
}

// ResendAPIRequest represents the request body for Resend API
type ResendAPIRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text,omitempty"`
}

// ResendAPIResponse represents the response from Resend API
type ResendAPIResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// NewResendClient creates a new Resend client instance
func NewResendClient(cfg *config.ResendConfig) *ResendClient {
	return &ResendClient{
		apiKey:    cfg.APIKey,
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendOTPEmail sends an OTP email using Resend API
// templateType can be: "register" or "reset_password"
func (rc *ResendClient) SendOTPEmail(toEmail, otp, templateType string) error {
	// Build the from address
	from := rc.fromEmail
	if rc.fromName != "" {
		from = fmt.Sprintf("%s <%s>", rc.fromName, rc.fromEmail)
	}

	// Generate email content based on template type
	subject, htmlContent, textContent := rc.generateOTPEmailContent(otp, templateType)

	// Prepare the request payload
	reqBody := ResendAPIRequest{
		From:    from,
		To:      []string{toEmail},
		Subject: subject,
		HTML:    htmlContent,
		Text:    textContent,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+rc.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := rc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, _ := io.ReadAll(resp.Body)

	// Check for non-200 status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Resend API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var resendResp ResendAPIResponse
	if err := json.Unmarshal(body, &resendResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Log success for debugging
	if resendResp.ID != "" {
		return nil
	}

	return nil
}

// generateOTPEmailContent generates email content based on template type
func (rc *ResendClient) generateOTPEmailContent(otp, templateType string) (subject, html, text string) {
	switch templateType {
	case "register":
		subject = "Kode OTP Registrasi Akun Anda"
		html = rc.generateRegisterOTPHTML(otp)
		text = rc.generateRegisterOTPText(otp)
	case "reset_password":
		subject = "Kode OTP Reset Password Anda"
		html = rc.generateResetPasswordOTPHTML(otp)
		text = rc.generateResetPasswordOTPText(otp)
	default:
		subject = "Kode OTP Anda"
		html = rc.generateDefaultOTPHTML(otp)
		text = rc.generateDefaultOTPText(otp)
	}
	return
}

// generateRegisterOTPHTML generates HTML email for registration OTP
func (rc *ResendClient) generateRegisterOTPHTML(otp string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Kode OTP Registrasi</title>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
		.content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
		.otp-code { background: white; border: 2px dashed #667eea; padding: 20px; text-align: center; font-size: 32px; font-weight: bold; letter-spacing: 5px; color: #667eea; margin: 20px 0; border-radius: 5px; }
		.footer { text-align: center; margin-top: 20px; font-size: 12px; color: #777; }
		.warning { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Selamat Datang!</h1>
			<p>Kode Verifikasi Akun Anda</p>
		</div>
		<div class="content">
			<p>Terima kasih telah mendaftar. Berikut adalah kode OTP (One-Time Password) untuk verifikasi akun Anda:</p>
			<div class="otp-code">%s</div>
			<p><strong>Berlaku selama 10 menit.</strong></p>
			<div class="warning">
				<strong>⚠️ Penting:</strong>
				<ul style="margin: 10px 0; padding-left: 20px;">
					<li>Jangan bagikan kode ini kepada siapa pun</li>
					<li>Tim kami tidak akan pernah meminta kode OTP Anda</li>
					<li>Kode ini bersifat rahasia dan hanya untuk Anda</li>
				</ul>
			</div>
			<p>Jika Anda tidak merasa melakukan pendaftaran, silakan abaikan email ini.</p>
		</div>
		<div class="footer">
			<p>Email ini dikirim secara otomatis, mohon jangan balas email ini.</p>
			<p>&copy; %d Invento Service. All rights reserved.</p>
		</div>
	</div>
</body>
</html>`, otp, time.Now().Year())
}

// generateRegisterOTPText generates plain text email for registration OTP
func (rc *ResendClient) generateRegisterOTPText(otp string) string {
	return fmt.Sprintf(`KODE OTP REGISTRASI AKUN

Selamat datang! Terima kasih telah mendaftar.

Kode OTP (One-Time Password) untuk verifikasi akun Anda:

%s

Berlaku selama 10 menit.

⚠️ PENTING:
- Jangan bagikan kode ini kepada siapa pun
- Tim kami tidak akan pernah meminta kode OTP Anda
- Kode ini bersifat rahasia dan hanya untuk Anda

Jika Anda tidak merasa melakukan pendaftaran, silakan abaikan email ini.

Email ini dikirim secara otomatis, mohon jangan balas email ini.

© %d Invento Service. All rights reserved.
`, otp, time.Now().Year())
}

// generateResetPasswordOTPHTML generates HTML email for reset password OTP
func (rc *ResendClient) generateResetPasswordOTPHTML(otp string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Kode OTP Reset Password</title>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
		.content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
		.otp-code { background: white; border: 2px dashed #f5576c; padding: 20px; text-align: center; font-size: 32px; font-weight: bold; letter-spacing: 5px; color: #f5576c; margin: 20px 0; border-radius: 5px; }
		.footer { text-align: center; margin-top: 20px; font-size: 12px; color: #777; }
		.warning { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; }
		.info { background: #d1ecf1; border-left: 4px solid #17a2b8; padding: 15px; margin: 20px 0; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Reset Password</h1>
			<p>Kode Verifikasi Reset Password</p>
		</div>
		<div class="content">
			<p>Kami menerima permintaan untuk mereset password akun Anda. Berikut adalah kode OTP (One-Time Password) untuk melanjutkan:</p>
			<div class="otp-code">%s</div>
			<p><strong>Berlaku selama 10 menit.</strong></p>
			<div class="warning">
				<strong>⚠️ Penting:</strong>
				<ul style="margin: 10px 0; padding-left: 20px;">
					<li>Jangan bagikan kode ini kepada siapa pun</li>
					<li>Tim kami tidak akan pernah meminta kode OTP Anda</li>
					<li>Kode ini bersifat rahasia dan hanya untuk Anda</li>
				</ul>
			</div>
			<div class="info">
				<strong>ℹ️ Informasi:</strong>
				<p style="margin: 10px 0;">Jika Anda tidak merasa meminta reset password, segera hubungi tim support kami dan amankan akun Anda.</p>
			</div>
		</div>
		<div class="footer">
			<p>Email ini dikirim secara otomatis, mohon jangan balas email ini.</p>
			<p>&copy; %d Invento Service. All rights reserved.</p>
		</div>
	</div>
</body>
</html>`, otp, time.Now().Year())
}

// generateResetPasswordOTPText generates plain text email for reset password OTP
func (rc *ResendClient) generateResetPasswordOTPText(otp string) string {
	return fmt.Sprintf(`KODE OTP RESET PASSWORD

Kami menerima permintaan untuk mereset password akun Anda.

Kode OTP (One-Time Password) untuk melanjutkan:

%s

Berlaku selama 10 menit.

⚠️ PENTING:
- Jangan bagikan kode ini kepada siapa pun
- Tim kami tidak akan pernah meminta kode OTP Anda
- Kode ini bersifat rahasia dan hanya untuk Anda

ℹ️ INFORMASI:
Jika Anda tidak merasa meminta reset password, segera hubungi tim support kami dan amankan akun Anda.

Email ini dikirim secara otomatis, mohon jangan balas email ini.

© %d Invento Service. All rights reserved.
`, otp, time.Now().Year())
}

// generateDefaultOTPHTML generates default HTML email for OTP
func (rc *ResendClient) generateDefaultOTPHTML(otp string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Kode OTP</title>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #4a90e2; color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
		.content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
		.otp-code { background: white; border: 2px dashed #4a90e2; padding: 20px; text-align: center; font-size: 32px; font-weight: bold; letter-spacing: 5px; color: #4a90e2; margin: 20px 0; border-radius: 5px; }
		.footer { text-align: center; margin-top: 20px; font-size: 12px; color: #777; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Kode OTP</h1>
		</div>
		<div class="content">
			<p>Berikut adalah kode OTP (One-Time Password) Anda:</p>
			<div class="otp-code">%s</div>
			<p><strong>Berlaku selama 10 menit.</strong></p>
			<p>Jangan bagikan kode ini kepada siapa pun.</p>
		</div>
		<div class="footer">
			<p>Email ini dikirim secara otomatis, mohon jangan balas email ini.</p>
			<p>&copy; %d Invento Service. All rights reserved.</p>
		</div>
	</div>
</body>
</html>`, otp, time.Now().Year())
}

// generateDefaultOTPText generates default plain text email for OTP
func (rc *ResendClient) generateDefaultOTPText(otp string) string {
	return fmt.Sprintf(`KODE OTP

Berikut adalah kode OTP (One-Time Password) Anda:

%s

Berlaku selama 10 menit.

Jangan bagikan kode ini kepada siapa pun.

Email ini dikirim secara otomatis, mohon jangan balas email ini.

© %d Invento Service. All rights reserved.
`, otp, time.Now().Year())
}
