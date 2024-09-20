package otp2fa

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"

	"github.com/skip2/go-qrcode"
)

func NewQrCodeUrl(
	urlKey string,
	accountTitle string,
	accountName string,
	issuer string,
	secret string,
) (string, string, []byte) {
	if dataParse, err := url.Parse(urlKey); err == nil {
		dataParse.Path = fmt.Sprintf("/%s:%s", accountTitle, accountName)
		dataParse.RawQuery = fmt.Sprintf("issuer=%s&secret=%s", issuer, secret)
		urlKey = dataParse.String()
	}
	var imageData string
	var pngData []byte
	var err error
	if pngData, err = qrcode.Encode(urlKey, qrcode.Medium, 256); err == nil {
		imageData = "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData)
	}
	return urlKey, imageData, pngData
}

func WriteFileQrCodeUrl(urlKey string, filepath string) error {
	return qrcode.WriteFile(urlKey, qrcode.Medium, 256, filepath)
}

func WriteFileByPngData(pngData []byte, filepath string) error {
	return os.WriteFile(filepath, pngData, 0o644)
}
