package main

import (
	"bytes"
	"html/template"
	"log"
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

func (m *Mail) SentSMTMessage(msg Message) error {
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
	formattedMesage, err := m.buildHTMLMessage(msg)
	if err != nil {
		log.Println(err)

		return err
	}

	plainMesage, err := m.buildPlanTextMessage(msg)
	if err != nil {
		log.Println(err)

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

	smptClient, err := server.Connect()
	if err != nil {
		log.Println(err)

		return err
	}
	email := mail.NewMSG()
	email.AddTo(msg.To).
		SetFrom(msg.From).
		SetSubject(msg.Subject)

	email.SetBody(mail.TextPlain, plainMesage)
	email.AddAlternative(mail.TextHTML, formattedMesage)
	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}
	err = email.Send(smptClient)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templetToRender := "./templates/mail.html.gohtml"

	t, err := template.New("email-html").ParseFiles(templetToRender)
	if err != nil {
		log.Println(err)
		return "", err
	}
	var tpl bytes.Buffer

	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		log.Println(err)
		return "", err
	}
	formattedMesage := tpl.String()
	formattedMesage, err = m.inlineCSS(formattedMesage)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return formattedMesage, nil
}

func (m *Mail) buildPlanTextMessage(msg Message) (string, error) {
	templetToRender := "./templates/mail.plain.gohtml"

	t, err := template.New("email-plain").ParseFiles(templetToRender)
	if err != nil {
		log.Println(err)
		return "", err
	}
	var tpl bytes.Buffer

	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		log.Println(err)
		return "", err
	}
	plainMesage := tpl.String()

	return plainMesage, nil
}

func (m *Mail) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}
	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}
	html, err := prem.Transform()
	if err != nil {
		return "", err
	}
	return html, nil
}

func (m *Mail) getEncryption(s string) mail.Encryption {
	switch s {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none", "":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSSLTLS
	}
}
