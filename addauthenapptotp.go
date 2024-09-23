package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
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
	var database, qrcode string
	func() {
		var newDataArgs []string
		for _, arg := range os.Args {
			switch true {
			case strings.HasPrefix(arg, "--database"),
				strings.HasPrefix(arg, "--qrcode"):
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
			case "--qrcode":
				if len(newDataArgs) > index {
					qrcode = strings.TrimSpace(newDataArgs[index+1])
				}
			}
		}
	}()
	if len(qrcode) == 0 {
		fmt.Println(
			`Usage: --database="totp.db" --qrcode="file-qrcode.png"`,
		)
		return
	}
	if len(database) == 0 {
		database = os.Getenv("TOTP_APP_DATABASE_FILENAME")
	}
	if len(database) == 0 {
		fmt.Println(
			`Usage: --database="totp.db" --qrcode="file-qrcode.png"`,
		)
		return
	}
	databasePath := os.Getenv("TOTP_APP_DATABASE_FOLDER") + database
	qrcodePath := os.Getenv("TOTP_APP_QRCODE_FOLDER") + qrcode
	data, err := os.ReadFile(qrcodePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	qrUrlKey, err := otp2fa.ReadQrCodeImage(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	if u, err := url.Parse(qrUrlKey); err == nil {
		issuer := u.Query().Get("issuer")
		secret := u.Query().Get("secret")
		if !strings.Contains(u.Path, ":") {
			u.Path = fmt.Sprintf("/%s:%s", issuer, strings.TrimPrefix(u.Path, "/"))
		}
		if before, after, found := strings.Cut(u.Path, ":"); found {
			accountTitle := strings.TrimPrefix(before, "/")
			accountName := after

			_ = os.MkdirAll(path.Dir(databasePath), 0o755)
			// Open a buckets database.
			bx, err := buckets.Open(databasePath)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer func(bx *buckets.DB) {
				_ = bx.Close()
			}(bx)

			totpDatabaseName := []byte("totpDatabase")
			totpDatabase, err := bx.New(totpDatabaseName)
			if err != nil {
				fmt.Println(err)
				return
			}

			secretDatabaseName := []byte("secretDatabase")
			secretDatabase, err := bx.New(secretDatabaseName)
			if err != nil {
				fmt.Println(err)
				return
			}

			beginDatabaseName := []byte("beginDatabase")
			beginDatabase, err := bx.New(beginDatabaseName)
			if err != nil {
				fmt.Println(err)
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
				firstTimeName := []byte("firstTime")
				if firstTime, err := beginDatabase.Get(firstTimeName); err == nil &&
					len(string(firstTime)) < 10 {
					_ = beginDatabase.Put(firstTimeName, timeNowData)
				}
			}()
			fmt.Printf("Added Account/Database:\n%s\n%s\n", string(secretKey), database)
		} else {
			fmt.Println("Path error", u.Path)
		}
	} else {
		fmt.Println("QR error", err)
	}
}

// go run addauthenapptotp.go --database="totp.db" --qrcode="admin-shopone-io-Bùi-Mạnh-manhavn-outlook-com.png"
