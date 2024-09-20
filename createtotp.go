package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
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
	var issuer, account, title string
	func() {
		var newDataArgs []string
		for _, arg := range os.Args {
			switch true {
			case strings.HasPrefix(arg, "--issuer"),
				strings.HasPrefix(arg, "--account"),
				strings.HasPrefix(arg, "--title"):
				for _, val := range strings.Split(arg, "=") {
					newDataArgs = append(newDataArgs, val)
				}
			default:
				newDataArgs = append(newDataArgs, arg)
			}
		}
		for index, arg := range newDataArgs {
			switch arg {
			case "--issuer":
				if len(newDataArgs) > index {
					issuer = strings.TrimSpace(newDataArgs[index+1])
				}
			case "--account":
				if len(newDataArgs) > index {
					account = strings.TrimSpace(newDataArgs[index+1])
				}
			case "--title":
				if len(newDataArgs) > index {
					title = strings.TrimSpace(newDataArgs[index+1])
				}
			}
		}
	}()
	if len(issuer) == 0 || len(account) == 0 || len(title) == 0 {
		fmt.Println(
			`Usage: --issuer="test.com" --account="hello@account.com" --title="Test Title"`,
		)
		return
	}

	key, urlKey, dataBase64PNG, pngData, err := otp2fa.NewTOTP(
		issuer,
		title,
		account,
		nil,
	)
	if err != nil {
		return
	}

	folderPath := os.Getenv("TOTP_APP_QRCODE_FOLDER")
	regexWordFilename := os.Getenv("TOTP_APP_REGEX_WORD_FILENAME")
	denimWordFilename := os.Getenv("TOTP_APP_DENIM_WORD_FILENAME")
	if regexWordFilename == "" {
		regexWordFilename = `[\p{L}\p{N}]+`
	}
	if denimWordFilename == "" {
		denimWordFilename = "-"
	}
	allTextData := fmt.Sprintf("%s-%s-%s", issuer, title, account)
	allString := regexp.MustCompile(regexWordFilename).FindAllString(allTextData, -1)
	filename := strings.Join(allString, denimWordFilename)
	_ = os.MkdirAll(folderPath, 0o755)

	func() {
		pngName := fmt.Sprintf("%s.png", filename)
		logName := fmt.Sprintf("%s.log", filename)
		qrFilename := fmt.Sprintf("%s%s", folderPath, pngName)
		logFilename := fmt.Sprintf("%s%s", folderPath, logName)
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

// go run createtotp.go --issuer="admin.shopone.io" --account="manhavn@outlook.com" --title="Bùi Mạnh"
