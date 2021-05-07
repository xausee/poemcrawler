package db

import (
	"PoemCrawler/models"
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"log"
	"os"
)

// Export 从数据库中导出所有诗歌，诗人和诗歌流派数据，格式为json lines
func Export() {
	ExportPoems()
	ExportPoets()
	ExportGenres()
}

// ExportPoems 从数据库中导出所有诗歌数据，格式为json lines
func ExportPoems() []models.Poem {
	db, err := models.NewDBManager()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.PoemCollection)

	var poems []models.Poem
	err = c.Find(bson.M{}).All(&poems)
	if err != nil {
		log.Println(err)
	}

	content := ""
	for _, poem := range poems {
		b, err := json.Marshal(poem)
		if err != nil {
			fmt.Printf("Error: %s", err)
		}
		content = content + string(b) + "\n"
	}

	writeFile("poems.json", content)

	return poems
}

// ExportPoets 从数据库中导出所有诗人数据，格式为json lines
func ExportPoets() []models.Poet {
	db, err := models.NewDBManager()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.PoetCollection)

	var poets []models.Poet
	err = c.Find(bson.M{}).All(&poets)
	if err != nil {
		log.Println(err)
	}

	content := ""
	for _, poet := range poets {
		b, err := json.Marshal(poet)
		if err != nil {
			fmt.Printf("Error: %s", err)
		}
		content = content + string(b) + "\n"
	}

	writeFile("poets.json", content)

	return poets
}

// ExportGenres 从数据库中导出所有诗歌流派数据，格式为json lines
func ExportGenres() []models.Genre {
	db, err := models.NewDBManager()
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.GenreCollection)

	var genres []models.Genre
	err = c.Find(bson.M{}).All(&genres)
	if err != nil {
		log.Println(err)
	}

	content := ""
	for _, genre := range genres {
		b, err := json.Marshal(genre)
		if err != nil {
			fmt.Printf("Error: %s", err)
		}
		content = content + string(b) + "\n"
	}

	writeFile("genres.json", content)

	return genres
}

func writeFile(fileName string, content string) {
	f, err := os.Create(fileName)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		log.Fatal(err)
	}
}
