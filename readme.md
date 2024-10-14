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
 go run addauthenapptotp.go --database="totp.db" --qrcode="test-com-Test-Title-hello-account-com.png"
```

- Update 2FA account from database

```shell
 go run totpupdateaccount.go --database="totp.db"
```

- Re-generate QR CODE from database to new folder (default folder: --output="qrcode")

```shell
 go run totpgenqrcode.go --database="totp.db" --output="new-qrcode"
```

- Show OTP Code from 2FA account

```shell
 go run totpshowcode.go --database="totp.db"
```

- Remove 2FA account from database

```shell
 go run totpremoveaccount.go --database="totp.db"
```

## Docker

- Build docker image

```shell
 sh build-docker.sh
```

- TEST script

```shell
 # add test folder
 mkdir test_otp2fa; cd test_otp2fa

 # CREATE global env
 mkdir env
 echo 'TOTP_APP_ENV=prod
TOTP_APP_RATE_COUNT=10
TOTP_APP_DATABASE_FILENAME=totp.db
TOTP_APP_QRCODE_FOLDER=qrcode/
TOTP_APP_DATABASE_FOLDER=database/
TOTP_APP_REGEX_WORD_FILENAME=[\p{L}\p{N}]+
TOTP_APP_DENIM_WORD_FILENAME=-
' > env/global.env

 # RUN container
 docker run -d --name otp2fa -v ${PWD}/env:/app/env -v ${PWD}/database:/app/database -v ${PWD}/qrcode:/app/qrcode -v ${PWD}/new-qrcode:/app/new-qrcode -it manhavn/otp2fa:v0.0.1
 docker exec otp2fa create --issuer="test.com" --account="hello@account.com" --title="Test Title"
 docker exec otp2fa load --database="totp.db" --qrcode="test-com-Test-Title-hello-account-com.png"
 docker exec -it otp2fa update --database="totp.db"
 docker exec -it otp2fa genqr --database="totp.db" --output="new-qrcode"
 docker exec -it otp2fa otp --database="totp.db"
 docker exec -it otp2fa remove --database="totp.db"
```
