package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-tty"

	"github.com/joho/godotenv"
	"github.com/variar/buckets"
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

	totpDatabaseName := []byte("totpDatabase")
	totpDatabase, err := bx.New(totpDatabaseName)
	if err != nil {
		fmt.Println(err)
		return
	}

	var firstTimeData []byte
	func() {
		beginDatabaseName := []byte("beginDatabase")
		beginDatabase, err := bx.New(beginDatabaseName)
		if err != nil {
			fmt.Println(err)
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

	mapAccountDeleteKey := map[string][]byte{}

	func() {
		rate, _ := strconv.Atoi(os.Getenv("TOTP_APP_RATE_COUNT"))
		if rate < 1 || rate > 100 {
			rate = 10
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
				accountTitle := arrTotp[2]
				arrTotp[2] = url.PathEscape(accountTitle)
				if len(arrTotp) > 2 {
					accountName := strings.Join(append(arrTotp[:3], strconv.Itoa(number)), "//")
					mapAccountDeleteKey[accountName] = item.Key
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

	getInputByStdIn := func(inputLabel, dataDefault, endCharInput string) string {
		fmt.Printf("\r\033[K%s%s", inputLabel, dataDefault)
		// Tạo tty để lắng nghe input từ bàn phím
		ttyObj, err := tty.Open()
		if err != nil {
			return dataDefault
		}
		defer func(ttyObj *tty.TTY) {
			_ = ttyObj.Close()
			fmt.Printf(endCharInput)
		}(ttyObj)

		var input []rune
		for _, ru := range dataDefault {
			input = append(input, ru)
		}

		// Lắng nghe từng phím bấm
		for {
			r, err := ttyObj.ReadRune()
			if err != nil {
				break
			}

			// Khi nhấn Enter thì dừng việc lắng nghe
			if r == '\r' || r == '\n' {
				if len(strings.TrimSpace(string(input))) > 0 {
					break
				} else {
					continue
				}
			}

			// Xử lý các phím khác
			switch r {
			case 127: // Phím Backspace
				if len(input) > 0 {
					input = input[:len(input)-1]
					fmt.Printf("\r%s%s\033[K", inputLabel, string(input)) // Xóa và in lại input
				}
			default:
				// Thêm ký tự vào input
				input = append(input, r)
				fmt.Printf("%s", string(r))
			}
		}
		return strings.TrimSpace(string(input))
	}

	func() {
		secretDatabaseName := []byte("secretDatabase")
		secretDatabase, err := bx.New(secretDatabaseName)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("\r\033[K\n\n\033[2A\r\033[K")
		for {
			input := getInputByStdIn("Enter the Account to update (q to quit): ", "", "\n")
			if input == "q" {
				os.Exit(0)
			}
			func() {
				arrTotp := strings.Split(input, "//")
				if len(arrTotp) < 3 {
					fmt.Printf("\033[1A\r\033[K")
					return
				}
				secretKey := []byte(strings.Join(arrTotp[:2], "//"))
				totpKey := mapAccountDeleteKey[input]
				if len(string(totpKey)) < 1 {
					fmt.Printf("\033[1A\r\033[K")
					return
				}
				secret, err := secretDatabase.Get(secretKey)
				if err != nil {
					fmt.Println(err)
					fmt.Printf("\033[2A\r\033[K")
					return
				}

				timeNowUnixNano := strconv.FormatInt(time.Now().UnixNano(), 10)
				timeNowData := []byte(timeNowUnixNano)
				issuer := getInputByStdIn("Change issuer: ", arrTotp[0], "\r\033[K")
				accountName := getInputByStdIn("Change account: ", arrTotp[1], "\r\033[K")
				if len(arrTotp) < 3 {
					arrTotp = append(arrTotp, arrTotp[0])
				}
				title, _ := url.PathUnescape(arrTotp[2])
				accountTitle := getInputByStdIn("Change title: ", title, "\r\033[K")
				accountTitle = url.PathEscape(accountTitle)

				newSecretKey := []byte(fmt.Sprintf(
					"%s//%s",
					issuer,
					accountName,
				))
				totpData := []byte(fmt.Sprintf(
					"%s//%s//%s//%s",
					issuer,
					accountName,
					accountTitle,
					timeNowUnixNano,
				))

				if string(secretKey) != string(newSecretKey) {
					if secretDatabase.Put(newSecretKey, secret) != nil {
						fmt.Printf("\033[1A\r\033[K")
						return
					}
					_ = secretDatabase.Delete(secretKey)
				}
				if totpDatabase.Put(timeNowData, totpData) != nil {
					fmt.Printf("\033[1A\r\033[K")
					return
				}
				_ = totpDatabase.Delete(totpKey)
				delete(mapAccountDeleteKey, input)

				fmt.Printf("Updated 2FA account: %s => %s\n", input, fmt.Sprintf(
					"%s//%s//%s",
					issuer,
					accountName,
					accountTitle,
				))
				fmt.Printf("\033[2A\r\033[KEnter the Account to update (q to quit): ")
			}()
		}
	}()
}

// go run totpupdateaccount.go --database="totp.db"
