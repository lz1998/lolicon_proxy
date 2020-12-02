package lolicon

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestSave(t *testing.T) {
	imageInfo := &ImageInfo{
		Title:  "title",
		Author: "asd",
		Pid:    123,
		P:      1234,
	}

	log.Infof("image info: %+v", imageInfo)
	if err := Db.Save(imageInfo).Error; err != nil {
		t.Error(err)
	}
}
