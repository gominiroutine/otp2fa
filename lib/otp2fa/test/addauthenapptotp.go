package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/variar/buckets"
	"root/lib/otp2fa"
)

func main() {
	databasePath := "data/database-totp.db"
	qrcodePath := "data/admin-shopone-io-Bùi-Mạnh-manhavn-outlook-com.png"
	if len(os.Args) >= 2 {
		qrcodePath = os.Args[1]
	}
	data, err := os.ReadFile(qrcodePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	if qrUrlKey, err := otp2fa.ReadQrCodeImage(data); err == nil {
		if u, err := url.Parse(qrUrlKey); err == nil {
			issuer := u.Query().Get("issuer")
			secret := u.Query().Get("secret")
			if before, after, found := strings.Cut(u.Path, ":"); found {
				accountTitle := strings.TrimPrefix(before, "/")
				accountName := after

				// Open a buckets database.
				bx, err := buckets.Open(databasePath)
				if err != nil {
					return
				}
				defer bx.Close()

				totpDatabaseName := []byte("totpDatabase")
				totpDatabase, err := bx.New(totpDatabaseName)
				if err != nil {
					return
				}

				secretDatabaseName := []byte("secretDatabase")
				secretDatabase, err := bx.New(secretDatabaseName)
				if err != nil {
					return
				}

				timeNowUnixNano := strconv.FormatInt(time.Now().UnixNano(), 10)
				timeNowData := []byte(timeNowUnixNano)
				secretKey := []byte(fmt.Sprintf(
					"%s//%s",
					issuer,
					accountName,
				))
				if dataSecret, err := secretDatabase.Get(secretKey); err == nil &&
					len(string(dataSecret)) > 10 {
					fmt.Println("Secret key exists:", string(secretKey))
					return
				}
				if secretDatabase.Put(secretKey, []byte(secret)) != nil {
					return
				}
				totpData := []byte(fmt.Sprintf(
					"%s//%s//%s//%s",
					issuer,
					accountName,
					accountTitle,
					timeNowUnixNano,
				))

				if totpDatabase.Put(timeNowData, totpData) != nil {
					return
				}

				func() {
					beginDatabaseName := []byte("beginDatabase")
					beginDatabase, err := bx.New(beginDatabaseName)
					if err != nil {
						return
					}
					firstTimeName := []byte("firstTime")
					if firstTime, err := beginDatabase.Get(firstTimeName); err == nil &&
						len(string(firstTime)) < 10 {
						_ = beginDatabase.Put(firstTimeName, timeNowData)
					}
				}()
				fmt.Println("Added:", string(secretKey), timeNowUnixNano)
			}
		}
	}
}

// go run data/addauthenapptotp.go data/admin-shopone-io-Bùi-Mạnh-manhavn-outlook-com.png
