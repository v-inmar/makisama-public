package email_utils

import (
	"app/internal/constants"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type EmailRequest struct {
	From     EmailAddress   `json:"from"`
	To       []EmailAddress `json:"to"`
	Subject  string         `json:"subject"`
	Text     string         `json:"text,omitempty"`
	HTML     string         `json:"html,omitempty"`
	Category string         `json:"category,omitempty"`
}

func SendContactEmailToAdmin(contactId int64, message string) error {
	reqBody := EmailRequest{
		From: EmailAddress{
			Email: os.Getenv("FROM_EMAIL"),
			Name:  "Makisama Admin",
		},
		To: []EmailAddress{
			{
				Email: os.Getenv("FROM_EMAIL"), // send to admin(me)
			},
		},
		Subject: "Contact Message Received",
		Text: fmt.Sprintf(`
		Contact Us Received
		Contact ID: %d
		Message:
		"%s"
		`, contactId, message),
		HTML: fmt.Sprintf(`
		<html>
		<body>
			<p>Contact Us Received</p>
			<p>Contact ID: %d</p>
			<p>Message:</p>
			<small>%s</small>
		</body>
		</html>
		`, contactId, message),
		Category: "Contact Us",
	}
	return emailSend(reqBody)
}

func SendFeedbackReportEmailToAdmin(reason string, feedbackId, boardId, reportId int64) error {
	reqBody := EmailRequest{
		From: EmailAddress{
			Email: os.Getenv("FROM_EMAIL"),
			Name:  "Makisama Admin",
		},
		To: []EmailAddress{
			{
				Email: os.Getenv("FROM_EMAIL"), // send to admin(me)
			},
		},
		Subject: "Feedback Reported",
		Text: fmt.Sprintf(`
		Feedback Reported
		Reason: %s
		Report ID: %d
		Feedback ID: %d
		Board ID: %d
		`, reason, reportId, feedbackId, boardId),
		HTML: fmt.Sprintf(`
		<html>
		<body>
			<p>Feedback Reported</p>
			<p>Reason: %s</p>
			<p>Report ID: %d</p>
			<p>Feedback ID: %d</p>
			<p>Board ID: %d</p>
		</body>
		</html>
		`, reason, reportId, feedbackId, boardId),
		Category: "Feedback Reported",
	}
	return emailSend(reqBody)
}

func SendBoardReportEmailToAdmin(reason string, boardId, reportId int64) error {
	reqBody := EmailRequest{
		From: EmailAddress{
			Email: os.Getenv("FROM_EMAIL"),
			Name:  "Makisama Admin",
		},
		To: []EmailAddress{
			{
				Email: os.Getenv("FROM_EMAIL"), // send to admin(me)
			},
		},
		Subject: "Board Reported",
		Text: fmt.Sprintf(`
		Board Reported
		Reason: %s
		Report ID: %d
		Board ID: %d
		`, reason, reportId, boardId),
		HTML: fmt.Sprintf(`
		<html>
		<body>
			<p>Board Reported</p>
			<p>Reason: %s</p>
			<p>Report ID: %d</p>
			<p>Board ID: %d</p>
		</body>
		</html>
		`, reason, reportId, boardId),
		Category: "Board Reported",
	}
	return emailSend(reqBody)
}

func SendPasswordChangedEmail(to, name string) error {
	reqBody := EmailRequest{
		From: EmailAddress{
			Email: os.Getenv("FROM_EMAIL"),
			Name:  "Makisama Support",
		},
		To: []EmailAddress{
			{
				Email: to,
			},
		},
		Subject: "Password Changed",
		Text: fmt.Sprintf(`
		Hi, %s

		You successfully changed your password
		`, name),
		HTML: fmt.Sprintf(`
		<html>
		<body>
			<h3>Hi, %s</h3>
			<p>You successfully changed your password</p>
		</body>
		</html>
		`, name),
		Category: "Password Changed",
	}

	return emailSend(reqBody)
}

func SendForgotPasswordEmail(to, name, token string) error {
	resetUrl := fmt.Sprintf("%s/auth/reset_password?token=%s", os.Getenv("APP_BASE_URL"), token)

	reqBody := EmailRequest{
		From: EmailAddress{
			Email: os.Getenv("FROM_EMAIL"),
			Name:  "Makisama Support",
		},
		To: []EmailAddress{
			{
				Email: to,
			},
		},
		Subject: "Reset Password",
		Text: fmt.Sprintf(`
		Hi, %s

		Use the link below to reset your password.
		%s
		This link is valid for 1 hour.
		`, name, resetUrl),
		HTML: fmt.Sprintf(`
		<html>
		<body>
			<h3>Hi, %s</h3>
			<p>Reset your password by clicking the link below</p>
			<a href="%s">Reset Password</a>
			<small>This link is valid for 1 hour</small>
		</body>
		</html>
		`, name, resetUrl),
		Category: "Reset Password",
	}

	return emailSend(reqBody)
}

func SendReplaceEmail(to, name, token string) error {
	return SendNewConfirmEmail(to, name, token)
}

func SendNewConfirmEmail(to, name, token string) error {
	confirmUrl := fmt.Sprintf("%s/auth/confirm_email?token=%s", os.Getenv("APP_BASE_URL"), token)

	reqBody := EmailRequest{
		From: EmailAddress{
			Email: os.Getenv("FROM_EMAIL"),
			Name:  "Makisama Support",
		},
		To: []EmailAddress{
			{
				Email: to,
			},
		},
		Subject: "Action Required: Confirm your email",
		Text: fmt.Sprintf(`
		Hi, %s

		Confirm your email by using this link
		%s
		This link is valid for 24 hours
		`, name, confirmUrl),
		HTML: fmt.Sprintf(`
		<html>
		<body>
			<h3>Hi, %s</h3>
			<p>Confirm your email by clicking the link below</p>
			<a href="%s">Confirm Email</a>
			<small>This link is valid for 24 hours</small>
		</body>
		</html>
		`, name, confirmUrl),
		Category: "Request Confirm",
	}

	return emailSend(reqBody)
}

func SendConfirmEmail(to, name, token string) error {
	confirmUrl := fmt.Sprintf("%s/auth/confirm_email?token=%s", os.Getenv("APP_BASE_URL"), token)

	reqBody := EmailRequest{
		From: EmailAddress{
			Email: os.Getenv("FROM_EMAIL"),
			Name:  "Makisama Support",
		},
		To: []EmailAddress{
			{
				Email: to,
			},
		},
		Subject: "Action Required: Confirm your email",
		Text: fmt.Sprintf(`
		Hi, %s

		Welcome to Makisama
		Confirm your email by using this link
		%s
		This link is valid for 24 hours
		`, name, confirmUrl),
		HTML: fmt.Sprintf(`
		<html>
		<body>
			<h3>Hi, %s</h3>
			<p>Welcome to Makisama</p>
			<p>Confirm your email by clicking the link below</p>
			<a href="%s">Confirm Email</a>
			<small>This link is valid for 24 hours</small>
		</body>
		</html>
		`, name, confirmUrl),
		Category: "Confirm Email",
	}

	return emailSend(reqBody)
}

func emailSend(reqBody EmailRequest) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		constants.MailtrapUrl,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}

	bearer := fmt.Sprintf("Bearer %s", os.Getenv("MAILTRAP_API_TOKEN"))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := io.ReadAll(resp.Body); err != nil {
		return err
	}
	return nil
}
