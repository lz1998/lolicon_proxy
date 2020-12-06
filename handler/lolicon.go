package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lz1998/lolicon_proxy/config"
	"github.com/lz1998/lolicon_proxy/service/lolicon"
	"github.com/lz1998/lolicon_proxy/util"
	log "github.com/sirupsen/logrus"
)

func Lolicon(c *gin.Context) {
	r18 := c.Query("r18")
	keyword := c.Query("keyword")
	var u string
	if keyword == "" {
		log.Infof("request has no keyword, use cached image")
		url, err := lolicon.GetUnusedImageUrl(r18 == "1" || strings.ToLower(r18) == "true")
		if err != nil {
			c.String(http.StatusInternalServerError, "get saved image url error, err: %+v", err)
			return
		}
		u = url
	} else {
		log.Infof("request has keyword, call lolicon api, apikey: %+v, r18: %+v", config.Apikey, r18)
		resp, err := lolicon.CallLolicon(config.Apikey, r18, keyword, "1")
		if err != nil {
			log.Errorf("failed to call lolicon api, err: %+v", err)
			c.String(http.StatusInternalServerError, "failed to call lolicon api")
			return
		}

		if len(resp.Data) == 0 {
			log.Warnf("resp data length 0")
			c.String(http.StatusInternalServerError, "resp data length 0")
			return
		}
		u = resp.Data[0].URL
	}

	filename, err := util.DownloadIfNotExist(u)
	if err != nil {
		c.String(http.StatusInternalServerError, "download error")
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, "/static/"+filename)
}
