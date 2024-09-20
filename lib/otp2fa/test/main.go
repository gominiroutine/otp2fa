package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"root/lib/otp2fa"
)

func main() {
	// Generate a new key for the user
	accountTitle := "Bùi Mạnh"
	// accountTitle := "Bui Manh"
	accountName := "gabongdaplus@giaminhmedia.vn"
	issuer := "Brevo"
	key, urlKey, dataBase64PNG, pngData, err := otp2fa.NewTOTP(
		issuer,
		accountTitle,
		accountName,
		nil,
	)
	if err != nil {
		return
	}
	_ = otp2fa.WriteFileQrCodeUrl(urlKey, "data/new.png")
	_ = otp2fa.WriteFileByPngData(pngData, "data/new2.png")

	secret := key.Secret()
	func() {
		data, err := os.ReadFile("data/new.png")
		if err != nil {
			return
		}
		if qrUrlKey, err := otp2fa.ReadQrCodeImage(data); err == nil {
			if u, err := url.Parse(qrUrlKey); err == nil {
				secret = u.Query().Get("secret")
			}
		}
	}()

	fmt.Println("Secret:", secret)
	fmt.Println("Key URL:", urlKey)
	fmt.Println(dataBase64PNG)

	go func() {
		for range time.Tick(time.Second) {
			timeNow := time.Now()
			countdown := 30 - timeNow.Second()%30
			if token, err := otp2fa.GenerateCode(secret, timeNow); err == nil {
				fmt.Println("Current OTP:", token)
			}
			time.Sleep(time.Second * time.Duration(countdown))
		}
	}()

	for {
		// Simulate user input for verification
		var userInput string
		fmt.Print("Enter the OTP: ")
		_, _ = fmt.Scan(&userInput)
		if userInput == "q" {
			os.Exit(0)
		}

		// Verify the user's input against the generated token
		if otp2fa.Validate(userInput, secret) {
			fmt.Println("OTP is valid!")
		} else {
			fmt.Println("Invalid OTP.")
		}
	}
}
