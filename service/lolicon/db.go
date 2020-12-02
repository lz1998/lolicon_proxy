package lolicon

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var Db *gorm.DB

func init() {
	db, err := gorm.Open(sqlite.Open("lolicon.db"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	Db = db
	err = Db.AutoMigrate(&ImageInfo{})
	if err != nil {
		panic(err)
	}
}
