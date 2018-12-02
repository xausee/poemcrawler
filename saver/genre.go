package db

import (
	"PoemCrawler/models"
	"log"
	"strings"
	"time"

	"github.com/mozillazg/go-pinyin"
	"gopkg.in/mgo.v2/bson"
)

// SaveGenres 保存诗歌流派到数据库
func SaveGenres(genres []models.Genre) (int, error) {
	if len(genres) == 0 {
		return 0, nil
	}

	db, err := models.NewDBManager()
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.GenreCollection)

	total, e := c.Find(bson.M{}).Count()
	if e != nil {
		log.Println(e)
		return 0, err
	}

	n := 0
	for _, genre := range genres {
		genre.TimeStamp = time.Now().Format("2006-01-02 15:04:05")
		genre.LastUpdateTimeStamp = time.Now().Format("2006-01-02 15:04:05")

		genre.AlphabetIndex = strings.ToUpper(pinyin.LazyPinyin(genre.Name, pinyin.NewArgs())[0])[0:1]
		genre.N = total + n + 1
		err = c.Insert(genre)
		if err != nil {
			log.Println("保存诗歌流派失败：", genre.Name)
			continue
		}
		n++
		log.Println("保存诗歌流派成功：", genre.Name)
	}

	return n, err
}
