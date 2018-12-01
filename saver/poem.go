package db

import (
	"PoemCrawler/models"
	"PoemCrawler/util"
	"log"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// SavePoem 保存单首诗歌到数据库
func SavePoem(p models.Poem, poemType string) error {
	db, err := models.NewDBManager()
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.PoemCollection)

	p.ID = bson.NewObjectId().Hex()
	p.Type = poemType
	p.TimeStamp = time.Now().Format("2006-01-02 15:04:05")
	p.LastUpdateTimeStamp = time.Now().Format("2006-01-02 15:04:05")

	err = c.Insert(p)
	if err != nil {
		log.Println("保存诗歌失败：", p.Title)
		return err
	}
	log.Println("保存诗歌成功：", p.Title)

	return nil
}

// SavePoems 批量保存诗歌到数据库
func SavePoems(ps []models.Poem, poemType string) (int, error) {
	if len(ps) == 0 {
		return 0, nil
	}

	db, err := models.NewDBManager()
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.PoemCollection)

	total, e := c.Find(bson.M{"type": poemType}).Count()
	if e != nil {
		log.Println(e)
		return 0, e
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

	return n, err
}
