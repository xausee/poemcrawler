package db

import (
	"PoemCrawler/models"
	"log"
	"strings"
	"time"

	"github.com/mozillazg/go-pinyin"
)

// SavePoet 保存诗人信息
func SavePoet(p models.Poet) error {
	db, err := models.NewDBManager()
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.PoetCollection)

	p.AlphabetIndex = strings.ToUpper(pinyin.LazyPinyin(p.Name, pinyin.NewArgs())[0])[0:1]
	p.Avatar = ""
	p.TimeStamp = time.Now().Format("2006-01-02 15:04:05")
	p.LastUpdateTimeStamp = time.Now().Format("2006-01-02 15:04:05")

	err = c.Insert(p)
	if err != nil {
		log.Println("保存诗人信息失败：", p.Name)
		return err
	}
	log.Println("保存诗人信息成功：", p.Name)

	return nil
}
