package ht

import (
	"PoemCrawler/models"
	"PoemCrawler/saver"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
)

func ParseAllFailPage() {
	urls := db.GetAllFailPageUrl()
	ParsePages(urls)
}

func ParsePages(urls []string) {
	chrs := GetPoetChronology()
	gens := GetPoemGenres()
	for _, url := range urls {
		log.Println("访问：" + url)
		doc := GetDocument(url)
		ParseXianDaiShi(chrs, gens, doc)
		// 移除解析过的失败页面记录
		db.DeleteFailPage(url)
	}
}

func GetDocument(url string) *goquery.Document {
	doc, err := goquery.NewDocument(url)

	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return doc
}

func GetPoetChronology() map[string]string {
	doc := GetDocument("http://www.shiku.org/shiku/xs/index.htm")

	if doc == nil {
		return make(map[string]string)
	}

	c := NewXianDaiShi(doc)
	poetChronology := c.GetPoetChronology()

	return poetChronology
}

func GetPoemGenres() []models.Genre {
	doc := GetDocument("http://www.shiku.org/shiku/xs/indexlp.htm")
	if doc == nil {
		return make([]models.Genre, 0, 50)
	}

	c := NewXianDaiShi(doc)
	genres := c.GetGenre()

	return genres
}

func ParseXianDaiShi(poetChronology map[string]string, genres []models.Genre, doc *goquery.Document) {
	p := strings.TrimLeft(doc.Url.Path, "/")
	ps := strings.Split(p, "/")

	suffix := ps[len(ps)-1]
	if strings.Contains(suffix, "index") {
		return
	}

	var isPoemCollection = false

	poemType := models.XianDai.String()
	c := NewXianDaiShi(doc)

	if len(ps) == 4 {
		if ps[2] == "yeshibin" {
			// 诗集的情况，一个页面多首诗
			// 如：http://www.shiku.org/shiku/xs/yeshibin/yeshibin_ztz_1.htm
			c.ParsePoet()
			c.ParsePoems()
		} else {
			// 诗集的情况，一个页面一首诗
			// 如：http://www.shiku.org/shiku/xs/haizi/154.htm
			c.ParsePoetFromOnePageOfCollection()
			c.ParsePoemFromOnePageOfCollection()
		}
		isPoemCollection = true
	} else {
		c.ParsePoet()
		c.ParsePoems()
	}

	poet := *c.Poet
	poems := c.Poems

	// 给诗人年代字段赋值
	if _, ok := poetChronology[suffix]; ok {
		poet.Chronology = poetChronology[suffix]
	}

	// 处理诗人流派信息
	gs := make([]string, 0, 5)
	for _, genre := range genres {
		for _, poetaddress := range genre.PoetAddresses {
			if poetaddress.Name == poet.Name && poetaddress.UrlAddress == suffix {
				gs = append(gs, genre.Name)
			}
		}
	}
	poet.Genres = gs

	// 保存数据
	db.Save(true, isPoemCollection, poet, poemType, poems, doc.Url.String())
}
