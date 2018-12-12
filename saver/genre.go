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
func SaveGenres(genres []models.Genre) int {
	if len(genres) == 0 {
		return 0
	}

	db, err := models.NewDBManager()
	if err != nil {
		log.Println(err)
		return 0
	}

	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.GenreCollection)

	total, err := c.Find(bson.M{}).Count()
	if err != nil {
		log.Println(err)
		return 0
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

	return n
}

func IsGenresSaved() bool {
	db, err := models.NewDBManager()
	if err != nil {
		log.Println(err)
		return false
	}
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.GenreCollection)

	count, err := c.Find(bson.M{}).Count()
	if err != nil {
		log.Println(err)
		return false
	}

	return count > 0
}
