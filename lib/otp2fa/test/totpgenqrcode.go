package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/variar/buckets"
	"root/lib/otp2fa"
)

func main() {
	const (
		rate         = 10
		breakMessage = "limit rate exceeded"
	)

	folderPath := "data/"

	databasePath := "data/database-totp.db"
	// Open a buckets database.
	bx, err := buckets.Open(databasePath)
	if err != nil {
		return
	}
	defer bx.Close()

	var firstTimeData []byte
	func() {
		beginDatabaseName := []byte("beginDatabase")
		beginDatabase, err := bx.New(beginDatabaseName)
		if err != nil {
			return
		}
		firstTimeName := []byte("firstTime")
		if firstTime, err := beginDatabase.Get(firstTimeName); err == nil &&
			len(string(firstTime)) > 10 {
			firstTimeData = firstTime
		}
	}()
	if firstTimeData == nil {
		fmt.Println("Data not found: beginDatabase/firstTime")
		return
	}

	func() {
		totpDatabaseName := []byte("totpDatabase")
		totpDatabase, err := bx.New(totpDatabaseName)
		if err != nil {
			return
		}

		startKey := firstTimeData
		var endKey []byte
		number := 1
		fmt.Println("No		|		Title			|		Account")
		for string(startKey) != string(endKey) {
			count := 0
			_ = totpDatabase.Map(func(k, v []byte) error {
				endKey = k
				count++
				if float64(count) >= rate {
					return errors.New(breakMessage)
				}
				return nil
			})
			dataDoings, _ := totpDatabase.RangeItems(startKey, endKey)
			startKey = endKey
			if len(dataDoings) == 0 {
				break
			}
			for _, item := range dataDoings {
				dataTotp := string(item.Value)
				arrTotp := strings.Split(dataTotp, "//")
				accountTitle := arrTotp[2]
				arrTotp[2] = url.PathEscape(accountTitle)
				if len(arrTotp) > 2 {
					fmt.Println(
						number,
						"		",
						accountTitle,
						"				",
						strings.Join(arrTotp[:3], "//"),
					)
					number++
				}
			}
		}
	}()

	func() {
		secretDatabaseName := []byte("secretDatabase")
		secretDatabase, err := bx.New(secretDatabaseName)
		if err != nil {
			return
		}

		for {
			var input string
			fmt.Print("Enter the Account (q to quit): ")
			_, _ = fmt.Scan(&input)
			if input == "q" {
				os.Exit(0)
			}
			func() {
				arrTotp := strings.Split(strings.TrimSpace(input), "//")
				issuer := arrTotp[0]
				accountName := arrTotp[1]
				accountTitle, _ := url.PathUnescape(arrTotp[2])
				secretKey := []byte(strings.Join(arrTotp[:2], "//"))

				if dataSecret, err := secretDatabase.Get(secretKey); err == nil &&
					len(string(dataSecret)) > 3 {
					secret := string(dataSecret)

					if key, err := otp2fa.GenerateKey(issuer, accountName, dataSecret); err == nil {
						urlKey, dataBase64PNG, pngData := otp2fa.NewQrCodeUrl(
							key.URL(),
							accountTitle,
							accountName,
							issuer,
							secret,
						)

						fmt.Println(urlKey)
						fmt.Println()
						fmt.Println(key.Secret())
						fmt.Println()
						fmt.Println(dataBase64PNG)
						fmt.Println()
						allTextData := fmt.Sprintf("%s-%s-%s", issuer, accountTitle, accountName)
						allString := regexp.MustCompile(`[\p{L}\p{N}]+`).
							FindAllString(allTextData, -1)
						filename := fmt.Sprintf(
							"%s%s.png",
							folderPath,
							strings.Join(allString, "-"),
						)
						//_ = otp2fa.WriteFileQrCodeUrl(urlKey, filename)
						_ = otp2fa.WriteFileByPngData(pngData, filename)
					}
				}
			}()
		}
	}()
}

// go run data/totpgenqrcode.go
