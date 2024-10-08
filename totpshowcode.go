package main

import (
	"context"
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
		fmt.Printf("\r\033[KUsage: --database=\"totp.db\"\n")
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
		fmt.Printf("\r\033[KData not found: beginDatabase/firstTime\n")
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

		fmt.Printf("\r\033[KNo\t\t|\t\tTitle\t\t\t|\t\tAccount\n")
		func() {
			dataDoings, err := totpDatabase.Items()
			if err != nil {
				fmt.Println(err)
				return
			}
			number := 1
			for _, item := range dataDoings {
				dataTotp := string(item.Value)
				arrTotp := strings.Split(dataTotp, "//")
				if len(arrTotp) > 2 {
					title := arrTotp[2]
					fmt.Printf(
						"\r\033[K%d\t\t%s\t\t\t\t%s\n",
						number,
						title,
						strings.Join(append(arrTotp[:2], strconv.Itoa(number)), "//"),
					)
					number++
				}
			}
		}()
	}()

	func() {
		secretDatabaseName := []byte("secretDatabase")
		secretDatabase, err := bx.New(secretDatabaseName)
		if err != nil {
			return
		}

		mapRunning := map[string]context.CancelFunc{}
		fmt.Printf("\n\n\033[2A\r\033[KEnter the Account (q to quit): ")
		for {
			var input string
			_, _ = fmt.Scan(&input)
			input = strings.TrimSpace(input)
			if input == "q" {
				os.Exit(0)
			}
			fmt.Printf("\r\033[K%s\033[1A\r\033[KEnter the Account (q to quit): ", input)
			for secret, cancelFunc := range mapRunning {
				cancelFunc()
				delete(mapRunning, secret)
			}
			func() {
				slideInput := strings.Split(input, "//")
				if len(slideInput) < 2 {
					return
				}
				secretKey := []byte(strings.Join(slideInput[:2], "//"))
				if dataSecret, err := secretDatabase.Get(secretKey); err == nil &&
					len(string(dataSecret)) > 3 {
					secret := string(dataSecret)
					go func(secret string) {
						ctx, cancel := context.WithCancel(context.Background())
						mapRunning[secret] = cancel
						for range time.Tick(time.Second) {
							timeNow := time.Now()
							countdown := 30 - timeNow.Second()%30
							if ctx.Err() != nil {
								break
							} else if token, err := otp2fa.GenerateCode(secret, timeNow); err == nil {
								fmt.Printf(
									"\n\r\033[K%s OTP: %s refresh at %d second(s)\n",
									input,
									token,
									countdown,
								)
								fmt.Printf("\033[2A\rEnter the Account (q to quit): ")
							}
							// time.Sleep(time.Second * time.Duration(countdown))
						}
					}(secret)
				}
			}()
		}
	}()
}

// go run totpshowcode.go --database="totp.db"
