package main

import (
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lz1998/lolicon_proxy/config"
	"github.com/lz1998/lolicon_proxy/handler"
	"github.com/lz1998/lolicon_proxy/service/lolicon"
	"github.com/lz1998/lolicon_proxy/util"
	log "github.com/sirupsen/logrus"
)

func init() {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)
}

func main() {
	apikey := os.Getenv("LOLICON_APIKEY")
	if apikey != "" {
		log.Infof("load LOLICON_APIKEY from ENV, %+v", apikey)
		config.Apikey = apikey
	} else {
		log.Warnf("failed to read LOLICON_APIKEY from ENV. Config Url: /config?apikey=xxx&cache_count=xxx")
	}

	cacheCount := os.Getenv("CACHE_COUNT")
	if cacheCount != "" {
		count, err := strconv.ParseInt(cacheCount, 10, 64)
		if err != nil {
			log.Error("failed to get CACHE_COUNT from ENV, not int")
			time.Sleep(5 * time.Second)
			os.Exit(0)
		}
		log.Infof("load CACHE_COUNT from ENV, %+v", cacheCount)
		config.CacheCount = count
	}

	PORT := os.Getenv("PORT")
	if PORT != "" {
		log.Infof("load PORT from ENV, %+v", PORT)
		PORT = ":" + PORT
	} else {
		PORT = ":18000"
	}

	greedy := os.Getenv("GREEDY")
	if greedy == "1" {
		log.Infof("load GREEDY from ENV, %+v", greedy)
		config.Greedy = true
	}

	// 启动时检测是否足够
	util.SafeGo(func() {
		lolicon.CheckImageCount(false)
		lolicon.CheckImageCount(true)
	})

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/config", handler.Config)
	router.GET("/lolicon", handler.Lolicon)
	router.Static("/static", "./static")
	if err := router.Run(PORT); err != nil {
		panic(err)
	}
}
