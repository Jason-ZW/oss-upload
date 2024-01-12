package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	OSSVersion = "v0.0.2"
	OSSDate    = time.Now()

	flags = map[string]map[string]string{
		"OSS_ENDPOINT":       {"aliyun oss endpoint.": "oss-cn-beijing.aliyuncs.com"},
		"OSS_BUCKET":         {"aliyun oss bucket.": ""},
		"OSS_KEY":            {"aliyun oss key.": ""},
		"OSS_SECRET":         {"aliyun oss secret.": ""},
		"OSS_EXPIRE_SECONDS": {"aliyun oss public url expire seconds.": "0"},
		"OSS_URL_PUBLIC":     {"after uploading, an accessible URL is generated.": "false"},
		"FILE_PATH":          {"file path which needs to be uploaded.": ""},
		"TARGET_PATH":        {"file path which in OSS": ""},
	}
)

func init() {
	cli.VersionPrinter = versionPrinter
}

func beforeFunc(c *cli.Context) error {
	return nil
}

func versionPrinter(c *cli.Context) {
	if _, err := fmt.Fprintf(c.App.Writer, OSSVersion); err != nil {
		logrus.Error(err)
	}
}

func main() {
	app := cli.NewApp()
	app.Author = "Jason-ZW"
	app.Before = beforeFunc
	app.EnableBashCompletion = true
	app.Name = os.Args[0]
	app.Usage = fmt.Sprintf("control and configure smartcart(%s)", OSSDate)
	app.Version = OSSVersion
	app.Flags = generateFlags()
	app.Action = action
	if err := app.Run(os.Args); err != nil {
		logrus.Fatalln(err)
	}
}

func setEnvironments(c *cli.Context) error {
	for k := range flags {
		if err := os.Setenv(k, c.String(strings.ToLower(k))); err != nil {
			return err
		}
	}
	return nil
}

func generateFlags() []cli.Flag {
	fgs := make([]cli.Flag, 0)
	for key, value := range flags {
		for k, v := range value {
			f := cli.StringFlag{
				Name:   strings.ToLower(key),
				EnvVar: key,
				Usage:  k,
				Value:  v,
			}
			fgs = append(fgs, f)
		}
	}
	fgs = append(fgs, cli.HelpFlag)
	return fgs
}

func action(c *cli.Context) error {
	if err := setEnvironments(c); err != nil {
		return err
	}

	client, err := oss.New(os.Getenv("OSS_ENDPOINT"), os.Getenv("OSS_KEY"), os.Getenv("OSS_SECRET"))
	if err != nil {
		logrus.Fatalln(err)
	}

	bucket, err := client.Bucket(os.Getenv("OSS_BUCKET"))
	if err != nil {
		logrus.Fatalln(err)
	}

	filePath := os.Getenv("FILE_PATH")
	targetPath := os.Getenv("TARGET_PATH")
	fileName := filepath.Base(filePath)
	if targetPath != "" {
		fileName = targetPath + "/" + fileName
	}

	isPublic := os.Getenv("OSS_URL_PUBLIC")
	if strings.EqualFold(isPublic, "true") {
		expireSeconds, _ := strconv.ParseInt(os.Getenv("OSS_EXPIRE_SECONDS"), 10, 64)
		signedURL, err := bucket.SignURL(fileName, oss.HTTPPut, expireSeconds)
		if err != nil {
			logrus.Fatalln(err)
		}
		err = bucket.PutObjectFromFileWithURL(signedURL, filePath)
	} else {
		fmt.Println("fileName:", fileName, "filePath:", filePath)
		err = bucket.PutObjectFromFile(fileName, filePath)
	}
	if err != nil {
		logrus.Fatalln(err)
	}

	logrus.Infof("success upload the file: %s", fileName)

	return nil
}
