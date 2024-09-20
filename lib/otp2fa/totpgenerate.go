package otp2fa

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func NewTOTP(
	issuer string,
	accountTitle string,
	accountName string,
	secret []byte,
) (key *otp.Key, urlKey string, dataBase64PNG string, pngData []byte, err error) {
	key, err = GenerateKey(issuer, accountName, secret)
	if err == nil {
		urlKey, dataBase64PNG, pngData = NewQrCodeUrl(
			key.URL(),
			accountTitle,
			accountName,
			issuer,
			key.Secret(),
		)
	}
	return
}

func GenerateKey(issuer string, accountName string, secret []byte) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
		Secret:      secret,
	})
}

func GenerateCode(secret string, t time.Time) (string, error) {
	return totp.GenerateCode(secret, t)
}

func Validate(passcode string, secret string) bool {
	return totp.Validate(passcode, secret)
}
