package db

import (
	"PoemCrawler/models"
	"PoemCrawler/util"
	"log"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// SavePoems 批量保存诗歌到数据库
func SavePoems(ps []models.Poem, poemType string) int {
	if len(ps) == 0 {
		return 0
	}

	db, err := models.NewDBManager()
	if err != nil {
		log.Println(err.Error())
	}
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.PoemCollection)

	total, err := c.Find(bson.M{"type": poemType}).Count()
	if err != nil {
		log.Println(err.Error())
		return 0
	}

	n := 0
	for _, p := range ps {
		p.ID = bson.NewObjectId().Hex()
		p.Title = strings.TrimSpace(p.Title)
		p.Type = poemType
		p.TimeStamp = time.Now().Format("2006-01-02 15:04:05")
		p.LastUpdateTimeStamp = time.Now().Format("2006-01-02 15:04:05")
		p.Content = util.TrimLeftSpaceKeep(p.Content)
		p.Content = util.TrimRightSpace(p.Content)
		p.N = total + n + 1

		err = c.Insert(p)
		if err != nil {
			log.Println("保存诗歌失败：", p.Title)
			continue
		}
		n++
		log.Println("保存诗歌成功：", p.Title)
	}

	return n
}
