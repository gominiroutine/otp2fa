package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

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
		fmt.Println(
			`Usage (default --output="qrcode"): --database="totp.db" --output="new-qrcode"`,
		)
		return
	}
	_ = os.MkdirAll(output, 0o755)
	if len(database) == 0 {
		database = os.Getenv("TOTP_APP_DATABASE_FILENAME")
	}
	if len(database) == 0 {
		fmt.Println(
			`Usage (default --output="qrcode"): --database="totp.db" --output="new-qrcode"`,
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
			fmt.Println(err)
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
			fmt.Println(err)
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

							fmt.Printf("Created 2FA/QR:\n%s\n%s\n", logName, pngName)

							// fmt.Println(urlKey)
							// fmt.Println()
							// fmt.Println(key.Secret())
							// fmt.Println()
							// fmt.Println(dataBase64PNG)
							// fmt.Println()
						}()
					}
				}
			}()
		}
	}()
}

// go run totpgenqrcode.go --database="totp.db" --output="new-qrcode"
