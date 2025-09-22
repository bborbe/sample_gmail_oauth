// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	libsentry "github.com/bborbe/sentry"
	"github.com/bborbe/service"
	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	app := &application{}
	os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
	SentryDSN       string `required:"false" arg:"sentry-dsn" env:"SENTRY_DSN" usage:"SentryDSN" display:"length"`
	SentryProxy     string `required:"false" arg:"sentry-proxy" env:"SENTRY_PROXY" usage:"Sentry Proxy"`
	CredentialsPath string `required:"false" arg:"credentials" env:"CREDENTIALS_PATH" usage:"path to credentials.json" default:"credentials.json"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
	glog.V(2).Info("starting OAuth2 token obtainer")

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

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("unable to bind to localhost:0: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", port)
	config.RedirectURL = redirectURL

	authURL := config.AuthCodeURL("state-token",
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce)

	glog.V(2).Infof("Open the following link in your browser: %v", authURL)

	tokenCh := make(chan *oauth2.Token, 1)
	errorCh := make(chan error, 1)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code in request", http.StatusBadRequest)
			errorCh <- fmt.Errorf("no code in request")
			return
		}

		tok, err := config.Exchange(ctx, code)
		if err != nil {
			http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
			errorCh <- fmt.Errorf("token exchange failed: %w", err)
			return
		}

		glog.Info("Authorization successful!")
		fmt.Fprintf(w, "Authorization successful! You can close this window.")
		tokenCh <- tok
	})

	go func() {
		if err := http.Serve(listener, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errorCh <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	select {
	case tok := <-tokenCh:
		glog.V(2).Infof("Access Token: %s", tok.AccessToken)
		glog.V(2).Infof("Refresh Token: %s", tok.RefreshToken)
		return nil
	case err := <-errorCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
