package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/variar/buckets"
	"root/lib/otp2fa"
)

func main() {
	const (
		rate         = 10
		breakMessage = "limit rate exceeded"
	)

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
				if len(arrTotp) > 2 {
					fmt.Println(
						number,
						"		",
						arrTotp[2],
						"				",
						strings.Join(arrTotp[:2], "//"),
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

		mapRunning := map[string]context.CancelFunc{}
		for {
			var input string
			fmt.Print("Enter the Account (q to quit): ")
			_, _ = fmt.Scan(&input)
			if input == "q" {
				os.Exit(0)
			}
			for secret, cancelFunc := range mapRunning {
				cancelFunc()
				delete(mapRunning, secret)
			}
			func() {
				secretKey := []byte(strings.TrimSpace(input))
				if dataSecret, err := secretDatabase.Get(secretKey); err == nil &&
					len(string(dataSecret)) > 3 {
					secret := string(dataSecret)
					go func(secret string) {
						ctx, cancel := context.WithCancel(context.Background())
						mapRunning[secret] = cancel
						for range time.Tick(time.Second) {
							timeNow := time.Now()
							countdown := 30 - timeNow.Second()%30
							if token, err := otp2fa.GenerateCode(secret, timeNow); err == nil {
								fmt.Println(
									input,
									"Current OTP:",
									token,
									"refresh at",
									countdown,
									"second(s)",
								)
							}
							time.Sleep(time.Second * time.Duration(countdown))
							if ctx.Err() != nil {
								break
							}
						}
					}(secret)
				}
			}()
		}
	}()
}

// go run data/totpshowcode.go
