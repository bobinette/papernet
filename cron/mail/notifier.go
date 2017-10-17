package mail

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"

	"github.com/bobinette/papernet/clients/auth"
	"github.com/bobinette/papernet/cron"
	"github.com/bobinette/papernet/errors"
)

type MailNotifier struct {
	authClient *auth.Client
	cron       cron.Cron

	email    string
	password string
	server   string
	port     int
}

func NewNotifierFactory(authClient *auth.Client, email, password, server string, port int) cron.NotifierFactory {
	return func(cron cron.Cron) (cron.Notifier, error) {
		return &MailNotifier{
			authClient: authClient,
			cron:       cron,

			email:    email,
			password: password,
			server:   server,
			port:     port,
		}, nil
	}
}

func (n *MailNotifier) Notify(ctx context.Context, papers []cron.Paper) error {
	user, err := n.authClient.User(n.cron.UserID)
	if err != nil {
		return err
	} else if user.Email == "" {
		return errors.New(fmt.Sprintf("no email for user %d", n.cron.UserID))
	}

	// Set up authentication information.
	auth := smtp.PlainAuth(
		n.email,
		n.email,
		n.password,
		n.server,
	)

	body := struct {
		Papers  []cron.Paper
		Link    string
		Q       string
		Sources string
	}{
		Papers:  papers,
		Link:    "https://papernet.bobi.space/search",
		Q:       n.cron.Q,
		Sources: strings.Join(n.cron.Sources, ","),
	}
	return NewRequest([]string{user.Email}, n.email, "Your search got new results", n.server, n.port).
		ParseTemplate("cron/mail/template.html", body).
		SendEmail(auth)
}

// Taken from https://medium.com/@dhanushgopinath/sending-html-emails-using-templates-in-golang-9e953ca32f3d

type Request struct {
	from    string
	to      []string
	subject string

	body string

	server string
	port   int

	err error
}

func NewRequest(to []string, from, subject, server string, port int) *Request {
	return &Request{
		from:    from,
		to:      to,
		subject: subject,

		server: server,
		port:   port,
	}
}

func (r *Request) SendEmail(auth smtp.Auth) error {
	if r.err != nil {
		return r.err
	}

	from := fmt.Sprintf("From: \"Papernet\" <%s>\n", r.from)
	to := fmt.Sprintf("To: %s\n", strings.Join(r.to, " "))
	subject := "Subject: " + r.subject + "!\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(from + to + subject + mime + "\n" + r.body)
	addr := fmt.Sprintf("%s:%d", r.server, r.port)

	if err := smtp.SendMail(addr, auth, r.from, r.to, msg); err != nil {
		return err
	}
	return nil
}

func (r *Request) ParseTemplate(templateFileName string, data interface{}) *Request {
	t, err := template.New("mail").Parse(mailTemplate)
	if err != nil {
		r.err = err
		return r
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		r.err = err
		return r
	}

	r.body = buf.String()
	return r
}
