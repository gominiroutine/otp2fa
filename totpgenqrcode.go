package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mattn/go-tty"
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
	var database, output string
	func() {
		var newDataArgs []string
		for _, arg := range os.Args {
			switch true {
			case strings.HasPrefix(arg, "--database"),
				strings.HasPrefix(arg, "--output"):
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
			case "--output":
				if len(newDataArgs) > index {
					output = strings.TrimSpace(newDataArgs[index+1])
				}
			}
		}
	}()
	if len(output) == 0 {
		output = strings.TrimSuffix(os.Getenv("TOTP_APP_QRCODE_FOLDER"), "/")
		fmt.Printf(
			"\r\003[K(default --output=\"%s\")\n\r\003[KUsage: --database=\"totp.db\" --output=\"new-qrcode\"\n",
			output,
		)
	}
	_ = os.MkdirAll(output, 0o755)
	if len(database) == 0 {
		database = os.Getenv("TOTP_APP_DATABASE_FILENAME")
	}
	if len(database) == 0 {
		fmt.Printf(
			"\r\003[K(default --output=\"%s\")\n\r\003[KUsage: --database=\"totp.db\" --output=\"new-qrcode\"\n",
			output,
		)
		return
	}
	databasePath := os.Getenv("TOTP_APP_DATABASE_FOLDER") + database

	// Open a buckets database.
	bx, err := buckets.Open(databasePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer bx.Close()

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

	func() {
		rate, _ := strconv.Atoi(os.Getenv("TOTP_APP_RATE_COUNT"))
		if rate < 1 || rate > 100 {
			rate = 10
		}
		totpDatabaseName := []byte("totpDatabase")
		totpDatabase, err := bx.New(totpDatabaseName)
		if err != nil {
			fmt.Println(err)
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
				accountTitle := arrTotp[2]
				arrTotp[2] = url.PathEscape(accountTitle)
				if len(arrTotp) > 2 {
					fmt.Printf(
						"\r\033[K%d\t\t%s\t\t\t\t%s\n",
						number,
						accountTitle,
						strings.Join(append(arrTotp[:3], strconv.Itoa(number)), "//"),
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
			input := getInputByStdIn("\r\033[KEnter the Account (q to quit): ", "", "\n")
			if input == "q" {
				os.Exit(0)
			}
			func() {
				arrTotp := strings.Split(strings.TrimSpace(input), "//")
				if len(arrTotp) < 3 {
					fmt.Printf("\033[1A")
					return
				}
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

						allTextData := fmt.Sprintf("%s-%s-%s", issuer, accountTitle, accountName)
						allString := regexp.MustCompile(`[\p{L}\p{N}]+`).
							FindAllString(allTextData, -1)
						filename := strings.Join(allString, "-")
						folderPath := output

						func() {
							pngName := fmt.Sprintf("%s.png", filename)
							logName := fmt.Sprintf("%s.log", filename)
							qrFilename := fmt.Sprintf("%s/%s", folderPath, pngName)
							logFilename := fmt.Sprintf("%s/%s", folderPath, logName)
							logData := []byte(fmt.Sprintf(
								"%s\n%s\n%s\n\n%s\n",
								urlKey,
								key.Secret(),
								dataBase64PNG,
								qrFilename,
							))
							_ = os.WriteFile(logFilename, logData, 0o644)
							//_ = otp2fa.WriteFileQrCodeUrl(urlKey, qrFilename)
							_ = otp2fa.WriteFileByPngData(pngData, qrFilename)

							fmt.Printf(
								"\r\033[KCreated 2FA/QR: %s\n\r\033[K%s\n\r\033[K%s\n",
								input,
								logName,
								pngName,
							)
							fmt.Printf("\033[4A\r\033[K")

							// fmt.Println(urlKey)
							// fmt.Println()
							// fmt.Println(key.Secret())
							// fmt.Println()
							// fmt.Println(dataBase64PNG)
							// fmt.Println()
						}()
						return
					}
				}
				fmt.Printf("\033[1A")
			}()
		}
	}()
}

// go run totpgenqrcode.go --database="totp.db" --output="new-qrcode"
