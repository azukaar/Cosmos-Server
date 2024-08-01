package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

var Template = `From: %s
To: %s
Subject: %s
MIME-Version: 1.0
Content-Type: text/html; charset=UTF-8

<!DOCTYPE html>
<html>
<head>
<style>
		body {
				font-family: Arial, sans-serif;
				line-height: 1.6;
		}
		.container {
				max-width: 500px;
				margin: auto;
				margin-top: 30px;
				padding: 20px;
		}
		.header {
				text-align: center;
				padding-bottom: 10px;
		}
		.logo {
				width: 130px;
		}
		.content {
				background-color: #f4f4f4;
				padding: 20px;
				border-radius: 5px;
				border-top: 2px solid rgb(171, 71, 188);
		}
		.footer {
			padding-top: 20px;
			color: #999;
		}
		h1 {
			color: rgb(171, 71, 188);
		}
		.button {
			background: rgb(171, 71, 188);
			color: white !important;
			padding: 10px 20px;
			border-radius: 5px;
			display: inline-block;
			margin-top: 15px;
			text-decoration: none;
		}
		.button:hover {
			background: rgb(141, 41, 168);
		}
</style>
</head>
<body>
<div class="container">
		<div class="header">
				<img src="%s" alt="Logo" class="logo">
		</div>
		<div class="content">
				%s
		</div>
		<div class="footer">
			Sent from: %s
		</div>
</div>
</body>
</html>`

func IsEmailEnabled() bool {
	config := GetMainConfig()
	return config.EmailConfig.Enabled
}

func IsNotifyLoginEmailEnabled() bool {
	config := GetMainConfig()
	return config.EmailConfig.NotifyLogin
}

func SendEmail(recipients []string, subject string, body string) error {
	config := GetMainConfig()

	hostPort := config.EmailConfig.Host + ":" + config.EmailConfig.Port
	auth := smtp.PlainAuth("", config.EmailConfig.Username, config.EmailConfig.Password, config.EmailConfig.Host)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.EmailConfig.AllowInsecureTLS,
		ServerName:         config.EmailConfig.Host,
	}

	ServerURL := GetServerURL("")
	LogoURL := ServerURL + "logo"

	send := func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		c, err := smtp.Dial(addr)
		if err != nil {
			return err
		}
		defer c.Close()

		if config.EmailConfig.UseTLS {
			if err = c.StartTLS(tlsConfig); err != nil {
				return err
			}
		}

		if err = c.Auth(a); err != nil {
			return err
		}

		if err = c.Mail(from); err != nil {
			return err
		}
		for _, addr := range to {
			if err = c.Rcpt(addr); err != nil {
				return err
			}
		}

		w, err := c.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(msg)
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}

		return c.Quit()
	}

	msg := []byte(fmt.Sprintf(
		Template,
		config.EmailConfig.From,
		strings.Join(recipients, ","),
		subject,
		LogoURL,
		body,
		ServerURL,
	))

	TriggerEvent(
		"cosmos.email.send",
		"Email sent",
		"success",
		"",
		map[string]interface{}{
			"recipients": recipients,
			"subject":    subject,
		})

	return send(hostPort, auth, config.EmailConfig.From, recipients, msg)
}
