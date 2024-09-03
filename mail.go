package main

import (
	"fmt"
	"net/smtp"
)

type Sender struct {
	email string
	pass  string
}

func sendMail(subject string, content string, reciever string, creds Sender) error {
	msg := "From: " + creds.email + "\n" +
		"To: " + reciever + "\n" +
		"Subject: " + subject + "\n\n" +
		content

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", creds.email, creds.pass, "smtp.gmail.com"),
		creds.email, []string{reciever}, []byte(msg))

	if err != nil {
		fmt.Printf("smtp error: %s", err)
		return err
	}

	return nil
}
