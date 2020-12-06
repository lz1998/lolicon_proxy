package lolicon

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/lz1998/lolicon_proxy/config"
	"github.com/lz1998/lolicon_proxy/util"
	log "github.com/sirupsen/logrus"
)

type LoliconResp struct {
	Code        int         `json:"code"`
	Msg         string      `json:"msg"`
	Quota       int         `json:"quota"`
	QuotaMinTTL int         `json:"quota_min_ttl"`
	Count       int         `json:"count"`
	Data        []ImageInfo `json:"data"`
}

type ImageInfo struct {
	Id     int      `gorm:"primaryKey"`
	Pid    int      `json:"pid" gorm:"column:pid"`
	P      int      `json:"p" gorm:"column:p"`
	UID    int      `json:"uid" gorm:"column:uid"`
	Title  string   `json:"title" gorm:"column:title"`
	Author string   `json:"author" gorm:"column:author"`
	URL    string   `json:"url" gorm:"column:url"`
	R18    bool     `json:"r18" gorm:"column:r18"`
	Width  int      `json:"width" gorm:"column:width"`
	Height int      `json:"height" gorm:"column:height"`
	Used   int      `gorm:"used"`
	Tags   []string `json:"tags" gorm:"-"`
}

const (
	LoliconUrl = "https://api.lolicon.app/setu/"
)

var (
	CheckLock    sync.Mutex
	GetLock      sync.Mutex
	LastCallTime int64
	Quota        int
	QuotaMinTTL  int
)

// 调用api.lolicon.app获取图片信息
func CallLolicon(apikey string, r18 string, keyword string, num string) (*LoliconResp, error) {
	req, err := http.NewRequest("GET", LoliconUrl, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	if apikey == "" {
		return nil, fmt.Errorf("apikey is not set")
	}
	q.Add("apikey", apikey)
	if r18 != "" && r18 != "0" {
		q.Add("r18", r18)
	}
	if keyword != "" {
		q.Add("keyword", keyword)
	}
	if num != "" {
		q.Add("num", num)
	}
	q.Add("size1200", "1")
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		buffer := bytes.NewBuffer(body)
		r, _ := gzip.NewReader(buffer)
		defer r.Close()
		unCom, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		body = unCom
	}
	var loliconResp LoliconResp
	if err = json.Unmarshal(body, &loliconResp); err != nil {
		return nil, err
	}
	if loliconResp.Code != 0 {
		return nil, fmt.Errorf(loliconResp.Msg)
	}
	Quota = loliconResp.Quota
	QuotaMinTTL = loliconResp.QuotaMinTTL
	LastCallTime = time.Now().Unix()
	return &loliconResp, nil
}

func GetImageCount(r18 bool) int64 {
	var count int64
	Db.Model(&ImageInfo{}).Where("used = 0").Where("r18 = ?", r18).Count(&count)
	return count
}

func CheckImageCount(r18 bool) {
	CheckLock.Lock()
	defer CheckLock.Unlock()
	var count = GetImageCount(r18)
	log.Infof("check image count, r18: %+v, count: %+v", r18, count)
	if count < config.CacheCount {
		PrepareImage(r18)
	}
}

func GreedyMode() {
	util.SafeGo(func() {
		for {
			time.Sleep(600 * time.Second)
			if (Quota > 50 || LastCallTime+int64(QuotaMinTTL) < time.Now().Unix()) && len(util.UrlChan) < 5 && time.Now().Unix()-LastCallTime > 300 {
				log.Infof("greedy mode is on, prepare image, quota: %+v, downloadChannelLength: %+v, lastCallTime: %+v", Quota, len(util.UrlChan), LastCallTime)
				PrepareImage(false)
				PrepareImage(true)
			} else {
				log.Warnf("not download")
				if !(Quota > 50 || LastCallTime+int64(QuotaMinTTL) < time.Now().Unix()) {
					log.Infof("greedy mode is on, but quota is not enough, %+v", Quota)
				}
				if !(len(util.UrlChan) < 5) {
					log.Infof("greedy mode is on, but download channel is not empty, %+v", len(util.UrlChan))
				}
				if !(time.Now().Unix()-LastCallTime > 15) {
					log.Infof("greedy mode is on, but lastCallTime is %+v", LastCallTime)
				}
			}
		}
	})
}

func PrepareImage(r18 bool) {
	log.Infof("call lolicon api to get image, apikey: %+v, r18: %+v", config.Apikey, r18)
	resp, err := CallLolicon(config.Apikey, func() string {
		if r18 {
			return "1"
		} else {
			return "0"
		}
	}(), "", "10")
	if err != nil {
		log.Errorf("failed to call lolicon api, err: %+v", err)
		return
	}
	log.Infof("succeed to call lolicon api, resp: %+v", string(util.JsonMustMarshal(resp)))
	for _, imageInfo := range resp.Data {
		util.AddDownloadUrl(imageInfo.URL) // 自动下载
		// 存json
		if err := SaveImageJson(&imageInfo); err != nil {
			log.Errorf("failed to save image info in json, err: %+v", err)
		}
		// 存数据库
		if err := Db.Save(&imageInfo).Error; err != nil {
			log.Errorf("failed to save image info in db, err: %+v", err)
		}
	}
	count := GetImageCount(r18)
	log.Infof("after preparing image, count: %+v, r18: %+v", count, r18)
}

func SaveImageJson(imageInfo *ImageInfo) error {
	imageInfoJsonBytes := util.JsonMustMarshal(imageInfo)
	jsonFileName := path.Base(imageInfo.URL) + ".json"
	if _, err := os.Stat("./json"); os.IsNotExist(err) {
		err := os.MkdirAll("./json", 0777)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile("./json/"+jsonFileName, imageInfoJsonBytes, 0777)
}

func GetUnusedImageUrl(r18 bool) (string, error) {
	util.SafeGo(func() { // 每次请求前检测数量是否足够，如果不足请求
		CheckImageCount(r18)
	})
	GetLock.Lock()
	defer GetLock.Unlock()
	var imageInfo ImageInfo
	err := Db.Model(&ImageInfo{}).Where("used = 0").Where("r18 = ?", r18).First(&imageInfo).Error
	if err != nil {
		return "", err
	}
	imageInfo.Used += 1 // 标记为已使用
	Db.Save(imageInfo)
	return imageInfo.URL, nil
}

func GetUnusedImageUrls(r18 bool) ([]string, error) {
	var imageInfos []*ImageInfo
	err := Db.Model(&ImageInfo{}).Where("used = 0").Where("r18 = ?", r18).Find(&imageInfos).Error
	if err != nil {
		return nil, err
	}

	urls := make([]string, 0)
	for _, imageInfo := range imageInfos {
		urls = append(urls, imageInfo.URL)
	}
	return urls, nil
}
