package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-tty"

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

	listAccountData := map[string]string{}
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
					accountTitle := arrTotp[2]
					accountName := strings.Join(append(arrTotp[:2], strconv.Itoa(number)), "//")
					listAccountData[strconv.Itoa(number)] = accountName
					fmt.Printf(
						"\r\033[K%d\t\t%s\t\t\t\t%s\n",
						number,
						accountTitle,
						accountName,
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

		// Tạo tty để lắng nghe input từ bàn phím
		ttyObj, err := tty.Open()
		if err != nil {
			return
		}
		defer func(ttyObj *tty.TTY) {
			_ = ttyObj.Close()
		}(ttyObj)

		inputLabel := "\r\033[KEnter the Account (q to quit): "
		var dataInput string
		input := &dataInput

		for {
			for {
				r, err := ttyObj.ReadRune()
				if err != nil {
					break
				}

				if r == 0x1B {
					continue
				}

				// Khi nhấn Enter thì dừng việc lắng nghe
				if r == '\r' || r == '\n' {
					if len(strings.TrimSpace(*input)) > 0 {
						break
					} else {
						continue
					}
				}

				// Xử lý các phím khác
				switch r {
				case 127: // Phím Backspace
					if len(*input) > 0 {
						*input = (*input)[:len(*input)-1]
						fmt.Printf("\r%s%s\033[K", inputLabel, *input) // Xóa và in lại input
					}
				default:
					newData := string(r)
					*input += newData
					fmt.Print(newData)
				}
			}

			*input = strings.TrimSpace(*input)
			if *input == "q" {
				os.Exit(0)
			}
			fmt.Printf("\n\r\033[K%s\033[1A\r\033[KEnter the Account (q to quit): ", *input)
			for secret, cancelFunc := range mapRunning {
				cancelFunc()
				delete(mapRunning, secret)
			}
			func() {
				inputRunning := *input
				*input = ""
				if !strings.Contains(inputRunning, "//") {
					inputRunning = listAccountData[strings.TrimSpace(inputRunning)]
				}
				slideInput := strings.Split(inputRunning, "//")
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
						func() {
							timeNow := time.Now()
							countdown := 30 - timeNow.Second()%30
							if token, err := otp2fa.GenerateCode(secret, timeNow); err == nil {
								fmt.Printf(
									"\n\r\033[K%s OTP: %s refresh at %d second(s)\n%s",
									inputRunning,
									token,
									countdown,
									token,
								)
								fmt.Printf("\033[2A\rEnter the Account (q to quit): %s", *input)
							}
						}()
						for range time.Tick(time.Second) {
							timeNow := time.Now()
							countdown := 30 - timeNow.Second()%30
							if ctx.Err() != nil {
								break
							} else if token, err := otp2fa.GenerateCode(secret, timeNow); err == nil {
								if countdown == 30 {
									fmt.Printf(
										"\n\r\033[K%s OTP: %s refresh at %d second(s)\n%s",
										inputRunning,
										token,
										countdown,
										token,
									)
									fmt.Printf("\033[2A\rEnter the Account (q to quit): %s", *input)
								} else {
									fmt.Printf(
										"\n\r\033[K%s OTP: %s refresh at %d second(s)\n",
										inputRunning,
										token,
										countdown,
									)
									fmt.Printf("\033[2A\rEnter the Account (q to quit): %s", *input)
								}
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
