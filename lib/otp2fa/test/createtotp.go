package main

import (
	"fmt"
	"regexp"
	"strings"

	"root/lib/otp2fa"
)

func main() {
	accountTitle := "Bùi Mạnh"
	accountName := "manhavn@outlook.com"
	issuer := "admin.shopone.io"

	key, urlKey, dataBase64PNG, pngData, err := otp2fa.NewTOTP(
		issuer,
		accountTitle,
		accountName,
		nil,
	)
	if err != nil {
		return
	}

	fmt.Println(urlKey)
	fmt.Println()
	fmt.Println(key.Secret())
	fmt.Println()
	fmt.Println(dataBase64PNG)
	fmt.Println()

	folderPath := "data/"
	allTextData := fmt.Sprintf("%s-%s-%s", issuer, accountTitle, accountName)
	allString := regexp.MustCompile(`[\p{L}\p{N}]+`).FindAllString(allTextData, -1)
	filename := fmt.Sprintf("%s%s.png", folderPath, strings.Join(allString, "-"))
	//_ = otp2fa.WriteFileQrCodeUrl(urlKey, filename)
	_ = otp2fa.WriteFileByPngData(pngData, filename)
}

// go run data/createtotp.go
