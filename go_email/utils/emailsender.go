package utils

import (
	"crypto/tls"
	//"fmt"
	"os"
	"gopkg.in/gomail.v2"
	"github.com/joho/godotenv"
	"log"

)

func SendEmail(to string, subject string, body string) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	m := gomail.NewMessage()
	m.SetHeader("From", "gsacademia5404@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer(os.Getenv("MAIL_HOST"), 587, os.Getenv("MAIL_USER"), os.Getenv("SMTP_MAIL_PASS"))
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}