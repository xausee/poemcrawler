package db

import (
	"PoemCrawler/models"
	"log"

	"gopkg.in/mgo.v2/bson"
)

type FailPage struct {
	ID  string // ID号
	Url string // 网页地址
}

// SaveFailPage 保存抓取失败的页面到数据库
func SaveFailPage(url string) {
	db, err := models.NewDBManager()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.FailPage)

	f := FailPage{
		ID:  bson.NewObjectId().Hex(),
		Url: url,
	}

	err = c.Insert(f)
	if err != nil {
		log.Println(err)
	}
}

// GetAllFailPageUrl 获取所有失败的记录
func GetAllFailPageUrl() []string {
	db, err := models.NewDBManager()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.FailPage)

	var pages []FailPage
	err = c.Find(bson.M{}).All(&pages)
	if err != nil {
		log.Println(err)
	}

	var urls []string
	for _, page := range pages {
		urls = append(urls, page.Url)
	}

	return urls
}

// DeleteFailPage 删除记录
func DeleteFailPage(url string) {
	db, err := models.NewDBManager()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.FailPage)

	err = c.Remove(bson.M{"url":url})
	if err != nil {
		log.Println(err)
	}
}
