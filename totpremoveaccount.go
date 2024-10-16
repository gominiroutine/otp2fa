package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mattn/go-tty"
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

	listAccountData := map[string]string{}
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
					listAccountData[strconv.Itoa(number)] = accountName
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

		input := dataDefault
		// Lắng nghe từng phím bấm
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
				if len(strings.TrimSpace(input)) > 0 {
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
					fmt.Printf("\r%s%s\033[K", inputLabel, input) // Xóa và in lại input
				}
			default:
				newData := string(r)
				input += newData
				fmt.Print(newData)
			}
		}
		return strings.TrimSpace(input)
	}

	func() {
		secretDatabaseName := []byte("secretDatabase")
		secretDatabase, err := bx.New(secretDatabaseName)
		if err != nil {
			fmt.Println(err)
			return
		}

		for {
			input := getInputByStdIn("\r\033[KEnter the Account to delete (q to quit): ", "", "\n")
			if input == "q" {
				os.Exit(0)
			}
			func() {
				if !strings.Contains(input, "//") {
					input = listAccountData[strings.TrimSpace(input)]
				}
				arrTotp := strings.Split(input, "//")
				if len(arrTotp) < 2 || len(string(mapAccountDeleteKey[input])) < 1 {
					fmt.Printf(
						"\r\033[K%s\033[1A\r\033[KEnter the Account to delete (q to quit): ",
						input,
					)
					return
				}
				secretKey := []byte(strings.Join(arrTotp[:2], "//"))

				_ = secretDatabase.Delete(secretKey)
				_ = totpDatabase.Delete(mapAccountDeleteKey[input])
				delete(mapAccountDeleteKey, input)

				fmt.Printf("Deleted 2FA account: %s\n", input)
				fmt.Printf("\033[2A\r\033[KEnter the Account to delete (q to quit): ")
			}()
		}
	}()
}

// go run totpremoveaccount.go --database="totp.db"
