package handler

import (
	"github.com/lz1998/lolicon_proxy/service/lolicon"
	"github.com/lz1998/lolicon_proxy/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lz1998/lolicon_proxy/config"
)

func Config(c *gin.Context) {
	apikey := c.Query("apikey")
	if apikey != "" {
		config.Apikey = apikey
	}
	cacheCount := c.Query("cache_count")
	if cacheCount != "" {
		count, err := strconv.ParseInt(cacheCount, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "cache_count not int")
		}
		config.CacheCount = count
	}
	util.SafeGo(func() { // 每次请求前检测数量是否足够，如果不足请求
		lolicon.CheckImageCount(false)
		lolicon.CheckImageCount(true)
	})
	c.JSON(http.StatusOK, map[string]interface{}{
		"apikey":      config.Apikey,
		"cache_count": config.CacheCount,
	})
}
