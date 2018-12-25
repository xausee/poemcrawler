package db

import (
	"PoemCrawler/models"
	"log"

	"gopkg.in/mgo.v2/bson"
)

// Address 页面处理信息概览数据结构
type Address struct {
	ID      string // ID号
	Count   int    // 解析到的诗歌数量
	Message string // 出错信息
	URL     string // 网页地址
}

// SaveAddress 保存单首诗歌到数据库
func SaveAddress(addr, msg string, count int) {
	db, err := models.NewDBManager()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.Address)

	a := Address{
		ID:      bson.NewObjectId().Hex(),
		Count:   count,
		Message: msg,
		URL:     addr,
	}

	err = c.Insert(a)
	if err != nil {
		log.Println(err)
	}
}
