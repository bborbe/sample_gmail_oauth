// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"fmt"
	"net/smtp"
)

func NewOAuth2Auth(username, accessToken string) smtp.Auth {
	return &oauth2Auth{username, accessToken}
}

type oauth2Auth struct {
	username    string
	accessToken string
}

func (a *oauth2Auth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	authString := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.username, a.accessToken)
	return "XOAUTH2", []byte(authString), nil
}

func (a *oauth2Auth) Next(fromServer []byte, more bool) ([]byte, error) {
	return nil, nil
}
