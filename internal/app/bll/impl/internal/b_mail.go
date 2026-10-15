package internal

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-gin/internal/app/bll"
	"go-gin/internal/app/config"
)

func NewMail(cfg *config.Config) *Mail {
	return &Mail{config: cfg}
}

type Mail struct {
	config *config.Config
}

var _ bll.IMail = (*Mail)(nil)

func (a *Mail) SendVerification(ctx context.Context, email, token string) error {
	link := a.config.Site.Origin + "/verify-email?token=" + url.QueryEscape(token)
	return a.sendMail(ctx, email, "确认你的邮箱 · 日落的足迹", "点击链接完成邮箱验证", link)
}

func (a *Mail) SendPasswordReset(ctx context.Context, email, token string) error {
	link := a.config.Site.Origin + "/reset-password?token=" + url.QueryEscape(token)
	return a.sendMail(ctx, email, "重置密码 · 日落的足迹", "点击链接设置新密码", link)
}

func (a *Mail) sendMail(ctx context.Context, recipient, subject, heading, link string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if a.config.Mail.Mode == "log" {
		log.Printf("[开发邮件] %s · %s · %s", recipient, subject, link)
		return nil
	}
	from, err := mail.ParseAddress(a.config.Mail.From)
	if err != nil {
		return fmt.Errorf("mail.from is not a valid address: %w", err)
	}
	to, err := mail.ParseAddress(recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}
	message, err := makeMailMessage(from, to, subject, heading, link)
	if err != nil {
		return err
	}
	address := net.JoinHostPort(a.config.Mail.Host, strconv.Itoa(a.config.Mail.Port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	tlsConfig := &tls.Config{
		ServerName: a.config.Mail.Host,
		MinVersion: tls.VersionTLS12,
	}
	var connection net.Conn
	if a.config.Mail.Port == 465 {
		tlsDialer := &tls.Dialer{
			NetDialer: dialer,
			Config:    tlsConfig,
		}
		connection, err = tlsDialer.DialContext(ctx, "tcp", address)
	} else {
		connection, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return fmt.Errorf("SMTP connection: %w", err)
	}
	defer connection.Close()
	stopCancellation := context.AfterFunc(ctx, func() {
		_ = connection.Close()
	})
	defer stopCancellation()
	// Bound the entire SMTP exchange, including peers that accept TCP but stall.
	if err := connection.SetDeadline(time.Now().Add(20 * time.Second)); err != nil {
		return err
	}
	client, err := smtp.NewClient(connection, a.config.Mail.Host)
	if err != nil {
		return fmt.Errorf("SMTP greeting: %w", err)
	}
	defer client.Close()
	if a.config.Mail.Port != 465 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("SMTP STARTTLS: %w", err)
			}
		}
	}
	if a.config.Mail.User != "" && a.config.Mail.Password != "" {
		_, mechanisms := client.Extension("AUTH")
		var auth smtp.Auth
		if strings.Contains(strings.ToUpper(mechanisms), "PLAIN") {
			auth = smtp.PlainAuth("", a.config.Mail.User, a.config.Mail.Password, a.config.Mail.Host)
		} else if strings.Contains(strings.ToUpper(mechanisms), "LOGIN") {
			auth = &smtpLoginAuth{
				username: a.config.Mail.User,
				password: a.config.Mail.Password,
				host:     a.config.Mail.Host,
			}
		} else {
			return fmt.Errorf("SMTP server does not offer PLAIN or LOGIN authentication")
		}
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication: %w", err)
		}
	}
	if err := client.Mail(from.Address); err != nil {
		return fmt.Errorf("SMTP sender: %w", err)
	}
	if err := client.Rcpt(to.Address); err != nil {
		return fmt.Errorf("SMTP recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP data: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("SMTP message acceptance: %w", err)
	}
	// DATA's positive response means acceptance. A missing QUIT reply must not
	// roll back a registration whose verification message has already left.
	_ = client.Quit()
	return nil
}

func makeMailMessage(from, to *mail.Address, subject, heading, link string) ([]byte, error) {
	var message bytes.Buffer
	parts := multipart.NewWriter(&message)
	fmt.Fprintf(&message, "From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMessage-ID: <%s@%s>\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%q\r\n\r\n",
		from.String(), to.String(), mime.QEncoding.Encode("UTF-8", subject), time.Now().Format(time.RFC1123Z),
		newAuthID(), strings.SplitN(from.Address, "@", 2)[1], parts.Boundary())
	plainText := heading + "：" + link + "\n如果这不是你的操作，请忽略这封邮件。"
	htmlText := `<div style="font-family:system-ui,sans-serif;max-width:560px;margin:auto;padding:32px"><h1 style="font-size:24px">` + html.EscapeString(heading) +
		`</h1><p>这个链接将在一段时间后失效。</p><p><a href="` + html.EscapeString(link) +
		`" style="display:inline-block;padding:10px 18px;border-radius:8px;background:#b85b3d;color:#fff;text-decoration:none">继续</a></p><p style="color:#666;font-size:13px">如果这不是你的操作，请忽略这封邮件。</p></div>`
	contentParts := []struct {
		contentType string
		body        string
	}{
		{
			contentType: "text/plain; charset=UTF-8",
			body:        plainText,
		},
		{
			contentType: "text/html; charset=UTF-8",
			body:        htmlText,
		},
	}
	for _, part := range contentParts {
		writer, err := parts.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {part.contentType},
			"Content-Transfer-Encoding": {"base64"},
		})
		if err != nil {
			return nil, err
		}
		if err := writeMIMEBase64(writer, part.body); err != nil {
			return nil, err
		}
	}
	if err := parts.Close(); err != nil {
		return nil, err
	}
	return message.Bytes(), nil
}

func writeMIMEBase64(writer io.Writer, value string) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(value))
	for len(encoded) > 0 {
		length := len(encoded)
		if length > 76 {
			length = 76
		}
		if _, err := io.WriteString(writer, encoded[:length]+"\r\n"); err != nil {
			return err
		}
		encoded = encoded[length:]
	}
	return nil
}

type smtpLoginAuth struct {
	username string
	password string
	host     string
}

func (a *smtpLoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if server.Name != a.host {
		return "", nil, fmt.Errorf("unexpected SMTP server")
	}
	if !server.TLS && server.Name != "localhost" && net.ParseIP(server.Name) == nil {
		return "", nil, fmt.Errorf("SMTP credentials require TLS")
	}
	if ip := net.ParseIP(server.Name); !server.TLS && ip != nil && !ip.IsLoopback() {
		return "", nil, fmt.Errorf("SMTP credentials require TLS")
	}
	return "LOGIN", nil, nil
}

func (a *smtpLoginAuth) Next(challenge []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimSpace(string(challenge))) {
	case "username:", "user name:":
		return []byte(a.username), nil
	case "password:":
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected SMTP LOGIN challenge")
	}
}
