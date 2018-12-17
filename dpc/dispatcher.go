package dpc

import (
	"PoemCrawler/ht"
	"PoemCrawler/models"
	"PoemCrawler/saver"
	"log"
	"net/http"
	"strings"

	"regexp"

	"github.com/PuerkitoBio/gocrawl"
	"github.com/PuerkitoBio/goquery"
)

// 现代诗人年表
var poetChronology = make(map[string]string)

// 诗歌流派信息
var genres = make([]models.Genre, 0, 50)

// Dispatcher 分派器
type Dispatcher struct {
	ctx *gocrawl.URLContext
	res *http.Response
	doc *goquery.Document
}

// NewDispatcher 分派器对象构造方法
func NewDispatcher(ctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) *Dispatcher {
	return &Dispatcher{ctx: ctx, res: res, doc: doc}
}

func (d Dispatcher) DispatchToSouYun() {
	souyun := ht.NewSouYun(d.doc)
	souyun.Parse()
	// 保存数据
	db.Save(true, false, souyun.Poet, models.GuDian.String(), souyun.Poems, d.doc.Url.String())
}

// Dispatch 执行分派
func (d Dispatcher) DispatchToShiKu() {
	p := strings.TrimLeft(d.doc.Url.Path, "/")
	ps := strings.Split(p, "/")

	t := ps[1]

	suffix := ps[len(ps)-1]
	if strings.Contains(suffix, "index") &&
		d.doc.Url.String() != "http://www.shiku.org/shiku/xs/index.htm" &&
		d.doc.Url.String() != "http://www.shiku.org/shiku/xs/indexlp.htm" {
		return
	}

	var poet models.Poet
	var poems []models.Poem
	var isPoemCollection = false
	var hasPoet = false
	poemType := models.WeiZhi.String()

	switch t {
	case "xs":
		// 先获取诗人年表数据，作为全局数据来使用
		if len(poetChronology) == 0 {
			poetChronology = ht.GetPoetChronology()
		}

		// 获取诗歌流派数据
		if len(genres) == 0 {
			genres = ht.GetPoemGenres()
			if !db.IsGenresSaved() {
				db.SaveGenres(genres)
			}
		}

		ht.ParseXianDaiShi(poetChronology, genres, d.doc)
	case "gs":
		poemType = models.GuDian.String()
		c := ht.NewGuDianShi(d.ctx, d.res, d.doc)
		log.Println(ps)
		if len(ps) == 4 && ps[2] == "tangshi" {
			poems = c.GetTangShi()
			hasPoet = false
		} else if len(ps) == 4 && ps[2] == "songci" {
			poems = c.GetSongCi()
			poet = *c.Poet
			hasPoet = true
		} else if len(ps) == 3 && (ps[2] == "shijing.htm" || ps[2] == "daya.htm" || ps[2] == "xiaoya.htm" || ps[2] == "song.htm") {
			poems = c.GetShiJing()
			hasPoet = false
		} else if len(ps) == 3 && ps[2] == "tangdai.htm" {
			poems = c.GetTangShiSanBaiShou()
			hasPoet = false
		} else if len(ps) == 4 && ps[2] == "yuefu" {
			reg := regexp.MustCompile(`\d+\.htm`)
			if reg.Match([]byte(ps[3])) && ps[3] != "00.htm" && ps[3] != "000.htm" {
				poems = c.GetYueFu()
				poet = *c.Poet
				hasPoet = false // 一页包含多个诗人，需要另外处理
			}
		} else if len(ps) == 4 {
			reg := regexp.MustCompile(`\d+\.htm`)
			if reg.Match([]byte(ps[3])) && ps[3] != "00.htm" && ps[3] != "000.htm" {
				poems = c.GetYuanQu()
				poet = *c.Poet
				hasPoet = true
			}
		} else {
			poems = c.GetPoems()
			poet = *c.Poet
			hasPoet = true
			if len(ps) == 3 && ps[2] == "chuci.htm" {
				hasPoet = false
			}
		}
	case "ws":
		poemType = models.WaiWen.String()
		if ps[2] == "wg" {
			poemType = models.YiShi.String()
		}
		c := ht.NewGuoJiShi(d.ctx, d.res, d.doc)
		if len(ps) == 5 && ps[3] == "dante" {
			poems = c.GetDanDingShenQu()
			hasPoet = false
		} else if d.ctx.URL().String() == "http://www.shiku.org/shiku/ws/zg/tang.htm" {
			poems = c.GetTangShiSanBaiShou()
			hasPoet = false
		} else if d.ctx.URL().String() == "http://www.shiku.org/shiku/ws/zg/dufu.htm" {
			poems = c.GetDuFu()
			poet = *c.Poet
			hasPoet = true
		} else if len(ps) == 5 && strings.Contains(ps[3], "tagore") {
			poems = c.GetTaiGeErPoems()
			hasPoet = false
		} else if len(ps) == 4 && (ps[3] == "mallarme.htm" || ps[3] == "andrade.htm" || ps[3] == "transtromo.htm" ||
			ps[3] == "wordsworth.htm" || ps[3] == "dqtc.htm" || ps[3] == "baxter.htm") {
			poems = c.GetPoemsH2AndP()
			poet = *c.Poet
			hasPoet = true
		} else if len(ps) == 4 && ps[3] == "quyuan.htm" {
			poems = c.GetPoemOfQuYuan()
			poet = *c.Poet
			hasPoet = true
		} else if len(ps) == 4 && ps[3] == "shijing.htm" {
			poems = c.GetPoemsWsZgShiJing()
			poet = *c.Poet
			hasPoet = true
		} else if len(ps) == 4 && ps[2] == "ww" && ps[3] == "apollinaire.htm" {
			poems = c.GetPoemsApollinaire()
			poet = *c.Poet
			hasPoet = true
		} else if len(ps) == 5 {
			reg := regexp.MustCompile(`\d+\.htm`)
			if reg.Match([]byte(ps[4])) && ps[4] != "00.htm" && ps[4] != "000.htm" {
				poems = c.GetSinglePoem()
				poet = *c.Poet
				hasPoet = true
			}
		} else {
			poems = c.GetPoems()
			poet = *c.Poet
			hasPoet = true
		}
	}

	// 保存数据
	db.Save(hasPoet, isPoemCollection, poet, poemType, poems, d.doc.Url.String())
}
