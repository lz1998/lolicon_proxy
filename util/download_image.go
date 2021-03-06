package util

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
)

var UrlChan = make(chan string, 500)

func init() {
	rand.Seed(time.Now().Unix())
}

func StartDownload() {
	go func() {
		for u := range UrlChan {
			func() {
				defer func() {
					e := recover()
					if e != nil {
						log.Errorf("download image err recovered: %+v", e)
					}
				}()
				if _, err := DownloadIfNotExist(u); err != nil {
					log.Errorf("failed to download image url: %+v err: %+v", u, err)
				}
			}()
			log.Infof("downloadUrl channel length: %+v", len(UrlChan))
		}
	}()
}

func AddDownloadUrl(u string) {
	UrlChan <- u
}

func DownloadIfNotExist(u string) (string, error) {
	log.Infof("start download image url: %+v", u)
	fileName := GetMD5Hash(u) + path.Ext(u)
	filePath := "./static/" + fileName

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if _, err := os.Stat("./static"); os.IsNotExist(err) {
			err := os.MkdirAll("./static", 0777)
			if err != nil {
				return "", err
			}
		}

		b, err := GetBytes(u)
		if err != nil {
			return "", err
		}
		//b[len(b)-1] = byte(rand.Intn(120)) // 混淆md5
		err = ioutil.WriteFile(filePath, b, 0644)
		if err != nil {
			return "", err
		}
		log.Infof("succeed to download image, filename: %+v, url: %+v", fileName, u)
		time.Sleep(1 * time.Second) // 冷却时间
	} else {
		log.Infof("image exists, filename: %+v, url: %+v", fileName, u)
	}

	return fileName, nil
}
