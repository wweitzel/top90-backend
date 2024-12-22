package email

import (
	"fmt"
	"net/smtp"
	"os"
)

func Send(subject string, body string) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	from := "wesweitzel@gmail.com"
	password := os.Getenv("TOP90_EMAIL_PASSWORD")
	if password == "" {
		return fmt.Errorf("error: TOP90_EMAIL_PASSWORD is not set")
	}

	to := []string{"wesweitzel@gmail.com"}

	message := []byte(fmt.Sprintf("Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", subject, body))

	auth := smtp.PlainAuth("", from, password, smtpHost)
	addr := smtpHost + ":" + smtpPort

	err := smtp.SendMail(addr, auth, from, to, message)
	if err != nil {
		return fmt.Errorf("error sending email: %s", err)
	}

	return nil
}
