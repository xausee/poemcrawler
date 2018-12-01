package ht

import (
	"PoemCrawler/models"
	"github.com/PuerkitoBio/gocrawl"
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"strings"
)

type SouYun struct {
	uctx    *gocrawl.URLContext
	res     *http.Response
	doc     *goquery.Document
	dynasty string
	Poet    *models.Poet
	poems   []models.Poem
}

// NewSouYun 创建SouYun实例
func NewSouYun(uctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) *SouYun {
	poet := getPoet(uctx, doc)
	poet.ID = bson.NewObjectId().Hex()

	return &SouYun{
		uctx:    uctx,
		res:     res,
		doc:     doc,
		dynasty: getDynasty(uctx),
		Poet:    poet,
		poems:   make([]models.Poem, 0, 0),
	}
}

func getPoet(uctx *gocrawl.URLContext, doc *goquery.Document) *models.Poet {
	values, _ := url.ParseQuery(uctx.URL().String())
	name := values.Get("author")

	if name == "" {
		name = doc.Find("span[class=\"title\"]").Text()
		name = strings.TrimSpace(name)
	}

	p := &models.Poet{
		Name: name,
	}

	return p
}

func getDynasty(uctx *gocrawl.URLContext) string {
	values, _ := url.ParseQuery(uctx.URL().String())
	dynasty := values.Get("http://sou-yun.com/PoemIndex.aspx?dynasty")
	if dynasty != "" {
		switch dynasty {
		case "XianQin":
			dynasty = "先秦"
			break
		case "Qin":
			dynasty = "秦"
			break
		case "Han":
			dynasty = "汉"
			break
		case "WeiJin":
			dynasty = "魏晋"
			break
		case "NanBei":
			dynasty = "南北朝"
			break
		case "Sui":
			dynasty = "隋"
			break
		case "Tang":
			dynasty = "唐"
			break
		case "Song":
			dynasty = "宋"
			break
		case "Liao":
			dynasty = "辽"
			break
		case "Jin":
			dynasty = "金"
			break
		case "Yuan":
			dynasty = "元"
			break
		case "Ming":
			dynasty = "明"
			break
		case "Qing":
			dynasty = "清"
			break
		case "Jindai":
			dynasty = "近现代"
			break
		case "Dangdai":
			dynasty = "当代"
			break
		default:
			dynasty = "古代"
		}
	}
	return dynasty
}

func (t *SouYun) GetPoems() (poems []models.Poem) {
	t.doc.Find("body").Find("div").Each(func(i int, s *goquery.Selection) {
		id, exist := s.Attr("id")
		if exist {
			if strings.Contains(id, "item_") {
				str := s.Find("div[class=\"title\"]").Text()
				title := str

				if strings.Contains(str, "（") && strings.Contains(str, "·") && strings.Contains(str, "）") {
					arr := strings.Split(str, "（")
					title = arr[0]
					title = strings.TrimSpace(title)

					if t.Poet.Name == "" {
						rear := arr[1]
						index0 := strings.LastIndex(rear, "·")
						index1 := strings.LastIndex(rear, "）")
						if index0 != -1 && index1 != -1 && index0 < index1 {
							name := rear[index0:index1]
							name = strings.Replace(name, "·", "", -1)
							t.Poet.Name = name
						}
					}
				}

				content := s.Find("div[class=\"content\"]").Text()
				content = strings.Replace(content, "。", "。\n", -1)
				content = strings.Replace(content, "？", "？\n", -1)
				content = strings.Replace(content, "：", "：\n", -1)
				content = strings.Replace(content, "；", "；\n", -1)
				content = strings.Replace(content, "！", "！\n", -1)
				content = strings.Replace(content, "章）", "章）\n", -1)
				commentary := s.Find("span[class=\"comment\"]").Text()

				poem := models.Poem{
					AuthorID:   t.Poet.ID,
					Author:     t.Poet.Name,
					Dynasty:    t.dynasty,
					Source:     t.uctx.URL().String(),
					Title:      title,
					Content:    content,
					Commentary: commentary,
				}
				poems = append(poems, poem)
			}
		}
	})

	return
}
