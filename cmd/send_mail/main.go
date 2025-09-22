// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"net/smtp"
	"os"

	"github.com/bborbe/sample_golang_gmail/pkg"
	libsentry "github.com/bborbe/sentry"
	"github.com/bborbe/service"
	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	smtpServer = "smtp.gmail.com"
	smtpPort   = "587"
)

func main() {
	app := &application{}
	os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
	SentryDSN       string `required:"false" arg:"sentry-dsn" env:"SENTRY_DSN" usage:"SentryDSN" display:"length"`
	SentryProxy     string `required:"false" arg:"sentry-proxy" env:"SENTRY_PROXY" usage:"Sentry Proxy"`
	CredentialsPath string `required:"false" arg:"credentials" env:"CREDENTIALS_PATH" usage:"path to credentials.json" default:"credentials.json"`
	RefreshToken    string `required:"true" arg:"refresh-token" env:"REFRESH_TOKEN" usage:"OAuth2 refresh token"`
	FromEmail       string `required:"true" arg:"from" env:"FROM_EMAIL" usage:"sender email address"`
	ToEmail         string `required:"true" arg:"to" env:"TO_EMAIL" usage:"recipient email address"`
	Subject         string `required:"false" arg:"subject" env:"SUBJECT" usage:"email subject" default:"Test Email"`
	Body            string `required:"false" arg:"body" env:"BODY" usage:"email body" default:"This is a test email sent via Gmail SMTP + OAuth2 in Go!"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
	glog.V(2).Info("starting email sender")

	b, err := os.ReadFile(a.CredentialsPath)
	if err != nil {
		return fmt.Errorf("unable to read credentials.json: %w", err)
	}

	scopes := []string{
		"https://mail.google.com/",
		//"https://www.googleapis.com/auth/gmail.send",
	}

	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return fmt.Errorf("unable to parse credentials.json: %w", err)
	}

	token := &oauth2.Token{
		RefreshToken: a.RefreshToken,
	}
	tokenSource := config.TokenSource(ctx, token)

	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("unable to get access token: %w", err)
	}
	glog.V(2).Info("got new token")

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", a.ToEmail, a.Subject, a.Body))

	auth := pkg.NewOAuth2Auth(a.FromEmail, newToken.AccessToken)
	addr := smtpServer + ":" + smtpPort

	if err := smtp.SendMail(addr, auth, a.FromEmail, []string{a.ToEmail}, msg); err != nil {
		return fmt.Errorf("error sending mail: %w", err)
	}

	glog.V(1).Infof("email sent successfully from %s to %s", a.FromEmail, a.ToEmail)
	return nil
}
