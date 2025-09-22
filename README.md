# Gmail send mail with Go

## Create OAuth client

https://console.cloud.google.com/auth/clients/

* Create Client
* Select Desktop
* Give it any name
* Create
* Download JSON to credentials.json
* OK

## Create Refresh Token

```bash
go run cmd/get_token/main.go \
-credentials="credentials.json" \
-v=2 
```

## Send mail

```
go run cmd/send_mail/main.go \
-credentials="credentials.json" \
-refresh-token="PLACE_REFRESH_TOKEN_HERE" \
-from="benjamin.borbe@gmail.com" \
-to="benjamin.borbe@gmail.com" \
-subject="Test Email from Makefile" \
-body="This is a test email sent via the Makefile" \
-v=2
```
