package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/variar/buckets"

	"root/lib/otp2fa"
)

func init() {
	func() {
		envLookup, found := os.LookupEnv("TOTP_APP_ENV_PATH_GLOBAL")
		if !found {
			envLookup = "env/global.env"
		}
		_ = godotenv.Load(envLookup)
	}()
}

func main() {
	var database string
	func() {
		var newDataArgs []string
		for _, arg := range os.Args {
			switch true {
			case strings.HasPrefix(arg, "--database"):
				for _, val := range strings.Split(arg, "=") {
					newDataArgs = append(newDataArgs, val)
				}
			default:
				newDataArgs = append(newDataArgs, arg)
			}
		}
		for index, arg := range newDataArgs {
			switch arg {
			case "--database":
				if len(newDataArgs) > index {
					database = strings.TrimSpace(newDataArgs[index+1])
				}
			}
		}
	}()
	if len(database) == 0 {
		database = os.Getenv("TOTP_APP_DATABASE_FILENAME")
	}
	if len(database) == 0 {
		fmt.Println(`Usage: --database="totp.db"`)
		return
	}
	databasePath := os.Getenv("TOTP_APP_DATABASE_FOLDER") + database

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
		rate, _ := strconv.Atoi(os.Getenv("TOTP_APP_RATE_COUNT"))
		if rate < 1 || rate > 100 {
			rate = 10
		}
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
				if count >= rate {
					return errors.New("limit rate exceeded")
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
		fmt.Println("Enter the Account (q to quit): ")
		for {
			var input string
			_, _ = fmt.Scan(&input)
			if input == "q" {
				os.Exit(0)
			}
			for secret, cancelFunc := range mapRunning {
				cancelFunc()
				delete(mapRunning, secret)
				fmt.Printf("\033[1A")
				fmt.Printf("\033[K")
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
								fmt.Printf("\033[1A")
								fmt.Printf("\033[K")
								fmt.Printf(
									"\r%s OTP: %s refresh at %d second(s)\n",
									input,
									token,
									countdown,
								)
							}
							// time.Sleep(time.Second * time.Duration(countdown))
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

// go run totpshowcode.go --database="totp.db"
