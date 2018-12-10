package db

import (
	"PoemCrawler/models"

	"gopkg.in/mgo.v2/bson"
)

type FailPage struct {
	ID  string // ID号
	Url string // 网页地址
}

// SaveFailPage 保存抓取失败的页面到数据库
func SaveFailPage(url string) error {
	db, err := models.NewDBManager()
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.FailPage)

	f := FailPage{
		ID:  bson.NewObjectId().Hex(),
		Url: url,
	}

	err = c.Insert(f)
	if err != nil {
		return err
	}

	return nil
}
