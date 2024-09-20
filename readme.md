# OTP2FA

## Config

- Config file: `env/global.env`

```dotenv
TOTP_APP_ENV=prod
TOTP_APP_RATE_COUNT=10

TOTP_APP_DATABASE_FILENAME=totp.db
TOTP_APP_QRCODE_FOLDER=qrcode/
TOTP_APP_DATABASE_FOLDER=database/
TOTP_APP_REGEX_WORD_FILENAME=[\p{L}\p{N}]+
TOTP_APP_DENIM_WORD_FILENAME=-
```

## APP Shell Script

- Create New 2FA

```shell
 go run createtotp.go --issuer="test.com" --account="hello@account.com" --title="Test Title"
```

- Scan QR CODE save 2FA account to database

```shell
 go run addauthenapptotp.go --database="totp.db" --qrcode="file-qrcode.png"
```

- Re-generate QR CODE from database to new folder (default folder: --output="qrcode")

```shell
 go run totpgenqrcode.go --database="totp.db" --output="new-qrcode"
```

- Show OTP Code from 2FA account

```shell
 go run totpshowcode.go --database="totp.db"
```
