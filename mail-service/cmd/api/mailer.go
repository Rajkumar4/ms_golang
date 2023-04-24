package main

import (
	"bytes"
	"log"
	"text/template"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	UserName    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}

func (m *Mail) SendSMTPMessage(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	data := map[string]any{
		"message": msg.Data,
	}

	msg.DataMap = data
	formatedMessage, err := m.bulildHTMLMessage(msg)
	if err != nil {
		log.Printf("Failed to get formated string: %s", err.Error())
		return err
	}
	plainTextString, err := m.bulildPalinTextMessage(msg)
	if err != nil {
		log.Printf("Failed to get plain string: %s", err.Error())
		return err
	}
	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.UserName
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtp, err := server.Connect()
	if err != nil {
		log.Printf("failed to start server: %s", err.Error())
		return err
	}

	email := mail.NewMSG()
	log.Printf("Check to add %s", msg.To)
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject).
		SetBody(mail.TextPlain, plainTextString).
		AddAlternative(mail.TextHTML, formatedMessage)
	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}
	err = email.Send(smtp)
	if err != nil {
		log.Printf("Failed to send mail:%s", err.Error())
		return err
	}
	return nil
}

func (m *Mail) bulildPalinTextMessage(msg Message) (string, error) {
	templetToRender := "./templates/mail.plain.gohtml"
	templates, err := template.New("email-plain").Parse(templetToRender)
	if err != nil {
		log.Printf("Failed to parse templates: %s", err.Error())
		return "", err
	}

	var tpl bytes.Buffer
	if err = templates.Execute(&tpl, msg.Data); err != nil {
		log.Printf("Failed to execute templates: %s", err.Error())
		return "", err
	}

	plainTextString := tpl.String()
	plainTextString, err = m.inlineCss(plainTextString)
	return plainTextString, nil
}

func (m *Mail) bulildHTMLMessage(msg Message) (string, error) {
	templetToRender := "./templates/mail.html.gohtml"

	templates, err := template.New("email-html").Parse(templetToRender)
	if err != nil {
		log.Printf("Failed to parse templates: %s", err.Error())
		return "", err
	}

	var tpl bytes.Buffer

	if err = templates.Execute(&tpl, msg.Data); err != nil {
		log.Printf("Failed to execute html template: %s", err.Error())
		return "", err
	}
	formatedString := tpl.String()
	formatedString, err = m.inlineCss(formatedString)
	return formatedString, nil
}

func (m *Mail) inlineCss(c string) (string, error) {
	opts := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}
	prem, err := premailer.NewPremailerFromString(c, &opts)
	if err != nil {
		log.Printf("Failed to get new premailer: %s", err.Error())
		return "", err
	}
	html, err := prem.Transform()
	if err != nil {
		log.Printf("Failed to get new premailer: %s", err.Error())
		return "", err
	}
	return html, nil
}

func (m *Mail) getEncryption(encryption string) mail.Encryption {
	switch encryption {
	case "tls":
		return mail.EncryptionTLS
	case "ssl":
		return mail.EncryptionSSL
	case "", "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
