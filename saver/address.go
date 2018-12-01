package db

import (
	"PoemCrawler/models"

	"gopkg.in/mgo.v2/bson"
)

type Address struct {
	ID      string // ID号
	Count   int    // 解析到的诗歌数量
	Message string // 出错信息
	Url     string // 网页地址
}

// SaveAddress 保存单首诗歌到数据库
func SaveAddress(addr, msg string, count int) error {
	db, err := models.NewDBManager()
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.Address)

	a := Address{
		ID:      bson.NewObjectId().Hex(),
		Count:   count,
		Message: msg,
		Url:     addr,
	}

	err = c.Insert(a)
	if err != nil {
		return err
	}

	return nil
}
