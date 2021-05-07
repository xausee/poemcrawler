package ht

import (
	"PoemCrawler/models"
	"PoemCrawler/util"
	db "PoemCrawler/saver"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/mgo.v2/bson"
)

// 现代诗人年表
var poetChronology = make(map[string]string)

// 诗歌流派信息
var genres = make([]models.Genre, 0, 50)

func init() {
	// 先获取诗人年表数据，作为全局数据来使用
	if len(poetChronology) == 0 {
		doc := GetDocument("http://www.shiku.org/shiku/xs/index.htm")

		c := NewXianDaiShi(doc)
		poetChronology = c.getPoetChronology()
	}

	// 获取诗歌流派数据
	if len(genres) == 0 {
		doc := GetDocument("http://www.shiku.org/shiku/xs/indexlp.htm")

		c := NewXianDaiShi(doc)
		genres = c.getGenre()
		db.SaveGenres(genres)
	}
}

// ParseAllFailPage 解析所有失败的页面
//func ParseAllFailPage() {
// 	urls := db.GetAllFailPageURL()
// 	ParsePages(urls)
// }

// ParsePages 解析指定的页面
// func ParsePages(urls []string) {
// 	chrs := GetPoetChronology()
// 	gens := GetPoemGenres()
// 	for _, url := range urls {
// 		log.Println("访问：" + url)
// 		doc := GetDocument(url)
// 		c := NewXianDaiShi(doc)
// 		c.Parse(chrs, gens, doc)
// 		// 移除解析过的失败页面记录
// 		db.DeleteFailPage(url)
// 	}
// }

// GetDocument 获取单个页面原始document数据
func GetDocument(url string) *goquery.Document {
	doc, err := goquery.NewDocument(url)

	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return doc
}

// XianDaiShi 处理现代诗歌的类型
// 页面样例 http://www.shiku.org/shiku/xs/xuzhimo.htm
type XianDaiShi struct {
	doc              *goquery.Document
	Poet             models.Poet
	IsPoemCollection bool
	Poems            []models.Poem
}

// NewXianDaiShi 创建现代诗对象
func NewXianDaiShi(doc *goquery.Document) *XianDaiShi {
	return &XianDaiShi{
		doc:              doc,
		IsPoemCollection: false,
		Poems:            make([]models.Poem, 0, 0),
	}
}

// Parse 解析现代诗
func (t *XianDaiShi) Parse() {
	p := strings.TrimLeft(t.doc.Url.Path, "/")
	ps := strings.Split(p, "/")

	suffix := ps[len(ps)-1]
	if strings.Contains(suffix, "index") {
		return
	}

	if len(ps) == 4 {
		if ps[2] == "yeshibin" {
			// 诗集的情况，一个页面多首诗
			// 如：http://www.shiku.org/shiku/xs/yeshibin/yeshibin_ztz_1.htm
			t.parsePoet()
			t.parsePoems()
		} else {
			// 诗集的情况，一个页面一首诗
			// 如：http://www.shiku.org/shiku/xs/haizi/154.htm
			t.parsePoetFromOnePageOfCollection()
			t.parsePoemFromOnePageOfCollection()
		}
		t.IsPoemCollection = true
	} else {
		t.parsePoet()
		t.parsePoems()
	}

	// 给诗人年代字段赋值
	if _, ok := poetChronology[suffix]; ok {
		t.Poet.Chronology = poetChronology[suffix]
	}

	// 处理诗人流派信息
	gs := make([]string, 0, 5)
	for _, genre := range genres {
		for _, poetaddress := range genre.PoetAddresses {
			if poetaddress.Name == t.Poet.Name && poetaddress.UrlAddress == suffix {
				gs = append(gs, genre.Name)
			}
		}
	}
	t.Poet.Genres = gs
}

// getPoetChronology 获取诗人年表信息
func (t XianDaiShi) getPoetChronology() map[string]string {
	chronologyMap := make(map[string]string)

	t.doc.Find("body").Find("table").Each(func(i int, s *goquery.Selection) {
		if i == 2 {
			s.Find("tbody").Find("tr").Each(func(j int, s1 *goquery.Selection) {
				bytes := []byte(s1.Find("td").Contents().Not("div").Text())
				chronology := strings.TrimSpace(util.GBK2Unicode(bytes))
				chronology = strings.TrimSuffix(chronology, "：")
				s1.Find("td").Find("div").Find("ul").Find("a").Each(func(k int, s2 *goquery.Selection) {
					href, existHref := s2.Attr("href")
					if existHref {
						chronologyMap[href] = chronology
					}
				})
			})
		}
	})

	return chronologyMap
}

// getGenre 获取诗歌流派
func (t XianDaiShi) getGenre() []models.Genre {
	genres := make([]models.Genre, 0, 50)
	genreChronologyMap := make(map[string]string)

	t.doc.Find("body").Find("table").Each(func(i int, s *goquery.Selection) {
		if i == 0 { // 解析获取流派年代信息
			s.Find("li").Each(func(j int, s1 *goquery.Selection) {
				bytes := []byte(s1.Contents().Not("a").Text())
				chronology := strings.TrimSpace(util.GBK2Unicode(bytes))
				chronology = chronology[0:strings.Index(chronology, "：")]

				s1.Find("a").Each(func(k int, s2 *goquery.Selection) {
					href, existHref := s2.Attr("href")
					if existHref {
						href = href[strings.Index(href, "#"):]
						genreChronologyMap[href] = chronology
					}
				})
			})
		} else { // 解析获取流派详细信息
			bytes := []byte(s.Find("td").Find("h2").Text())
			genreName := util.GBK2Unicode(bytes)
			genreName = strings.Replace(genreName, "聽", "", -1)
			genreName = strings.TrimSpace(genreName)

			href, _ := s.Find("td").Find("h2").Find("a").Attr("href")

			bytes = []byte(s.Find("td").Contents().Not("h2").Not("a").Text())
			description := util.GBK2Unicode(bytes)
			description = strings.Replace(description, "聽", "", -1)
			description = strings.TrimSpace(description)

			pas := make([]models.PoetAddress, 0, 100)
			poetNames := ""
			s.Find("tr").Find("a").Each(func(j int, s1 *goquery.Selection) {
				href, existHref := s1.Attr("href")
				if existHref && !strings.Contains(href, "#") {
					bytes := []byte(s1.Text())
					poetName := strings.TrimSpace(util.GBK2Unicode(bytes))
					poetNames = poetNames + " " + poetName

					pa := models.PoetAddress{
						Name:       poetName,
						UrlAddress: href,
					}
					pas = append(pas, pa)
				}
			})

			description = description + poetNames
			genre := models.Genre{
				ID:            bson.NewObjectId().Hex(),
				Name:          genreName,
				Description:   description,
				Chronology:    genreChronologyMap[href],
				PoetAddresses: pas,
			}
			genres = append(genres, genre)
		}

	})

	return genres
}

// getFirstPoemTitleWithSep 获取页面上第一首诗的标题
func (t XianDaiShi) getFirstPoemTitleWithSep() string {
	titles := make([]string, 0, 0)
	var title string
	//has999999 := false
	t.doc.Find("body").Find("a").Each(func(i int, s *goquery.Selection) {
		href, existHref := s.Attr("href")
		if existHref {
			if strings.HasPrefix(href, "#") {
				gbkTitle := s.Text()
				titleBytes := []byte(gbkTitle)
				title = strings.TrimSpace(util.GBK2Unicode(titleBytes))
				titles = append(titles, title+ALinkTitleSep)
				s.AppendHtml(ALinkTitleSep)
				return
			}
		}

		//name, existName := s.Attr("name")
		//
		//if existName {
		//	if strings.Contains("0123456789", name[0:1]) {
		//		t.doc.Find("body").Find("a[name=\"" + name + "\"]").AppendHtml(sep)
		//	}
		//}
		//
		//if name == "999999" {
		//	has999999 = true
		//}
	})
	if len(titles) > 0 {
		return titles[0]
	}

	return ""
}

// parsePoet 解析网页获取诗人信息
func (t *XianDaiShi) parsePoet() models.Poet {
	gbkStr := t.doc.Find("title").Text()
	bytes := []byte(gbkStr)
	title := strings.TrimSpace(util.GBK2Unicode(bytes))

	var name string
	arr := strings.Split(title, "::")

	for _, v := range arr {
		if strings.Contains(v, "诗选") || strings.Contains(v, "诗集") {
			name = strings.TrimSpace(strings.Split(v, "诗选")[0])
			name = strings.TrimSpace(strings.Split(name, "诗集")[0])
		}
	}

	if name == "" {
		gbkStr = t.doc.Find("body").Find("h1").Text()
		bytes = []byte(gbkStr)
		title = strings.TrimSpace(util.GBK2Unicode(bytes))
		name = strings.TrimSpace(strings.Split(title, "诗选")[0])
		name = strings.TrimSpace(strings.Split(name, "诗集")[0])
	}

	ft := t.getFirstPoemTitleWithSep()
	if ft == "" {
		t.Poet = models.Poet{
			Name:   name,
			Intro:  "",
			Source: t.doc.Url.String(),
		}
	} else {
		gbkStr = t.doc.Find("body").Text()
		bytes = []byte(gbkStr)
		text := strings.TrimSpace(util.GBK2Unicode(bytes))
		text = strings.Replace(text, name+"诗选", "", 1)

		index := strings.Index(text, ft)
		if index > 0 {
			text = text[0:index]
		}

		intro := strings.TrimSpace(text)
		t.Poet = models.Poet{
			Name:   name,
			Intro:  intro,
			Source: t.doc.Url.String(),
		}
	}

	t.Poet.ID = bson.NewObjectId().Hex()

	return t.Poet
}

// parsePoetFromOnePageOfCollection 从诗集子页面获取诗人信息
func (t *XianDaiShi) parsePoetFromOnePageOfCollection() models.Poet {
	gbkStr := t.doc.Find("title").Text()
	bytes := []byte(gbkStr)
	title := strings.TrimSpace(util.GBK2Unicode(bytes))

	var name string
	arr := strings.Split(title, "::")

	for _, v := range arr {
		if strings.Contains(v, "诗选") || strings.Contains(v, "诗集") {
			name = strings.TrimSpace(strings.Split(v, "诗选")[0])
			name = strings.TrimSpace(strings.Split(name, "诗集")[0])
		}
	}

	// 标题里面不包括诗人名字的情况
	// 如：http://www.shiku.org/shiku/xs/guomoruo/guomr08.htm
	if name == "" {
		gbkStr = t.doc.Find("body").Find("a[href=\"index.htm\"]").Text()
		bytes = []byte(gbkStr)
		title = strings.TrimSpace(util.GBK2Unicode(bytes))
		name = strings.TrimSpace(strings.Split(title, "诗选")[0])
		name = strings.TrimSpace(strings.Split(name, "诗集")[0])
	}

	t.Poet = models.Poet{
		ID:     bson.NewObjectId().Hex(),
		Name:   name,
		Source: t.doc.Url.String(),
	}

	return t.Poet
}

// parsePoemsH2AndP 标题为h2标签，诗歌内容为h2标签后第二个标签内
// 例子页面 http://www.shiku.org/shiku/xs/mudan.htm        第二个标签为p
// 例子页面 http://www.shiku.org/shiku/xs/xuzhimo.htm      第二个标签为pre
func (t *XianDaiShi) parsePoemsH2AndP() []models.Poem {
	t.doc.Find("body").Find("h2").Each(func(i int, s *goquery.Selection) {
		gbkFullTitle := s.Text()
		fullTitleBytes := []byte(gbkFullTitle)
		fullTitle := strings.TrimSpace(util.GBK2Unicode(fullTitleBytes))
		title := strings.TrimSpace(strings.Split(fullTitle, " ")[0])
		subTitle := fullTitle[len(title):]

		gbkContent := s.Next().Next().Text()
		contentBytes := []byte(gbkContent)
		content := util.GBK2Unicode(contentBytes)

		poem := models.Poem{
			AuthorID: t.Poet.ID,
			Author:   t.Poet.Name,
			Source:   t.doc.Url.String(),
			Title:    title,
			SubTitle: subTitle,
			Content:  content,
		}

		t.Poems = append(t.Poems, poem)
	})

	return t.Poems
}

// parsePoemsPAndP 标题为p align="center" 标签， 诗歌内容为标题标签后第一个p标签内
// 例子页面 http://www.shiku.org/shiku/xs/guangweiran.htm
func (t *XianDaiShi) parsePoemsPAndP() []models.Poem {
	t.doc.Find("body").Find("p[align=\"center\"]").Each(func(i int, s *goquery.Selection) {
		gbkFullTitle := s.Text()
		fullTitleBytes := []byte(gbkFullTitle)
		fullTitle := strings.TrimSpace(util.GBK2Unicode(fullTitleBytes))
		title := strings.TrimSpace(strings.Split(fullTitle, " ")[0])
		subTitle := fullTitle[len(title):]

		gbkContent := s.Next().Text()
		contentBytes := []byte(gbkContent)
		content := util.GBK2Unicode(contentBytes)

		// 起始作者信息介绍里，p align="center" 标签内内容为空，忽略
		if title != "" {
			poem := models.Poem{
				AuthorID: t.Poet.ID,
				Author:   t.Poet.Name,
				Source:   t.doc.Url.String(),
				Title:    title,
				SubTitle: subTitle,
				Content:  content,
			}

			t.Poems = append(t.Poems, poem)
		}
	})

	return t.Poems
}

// parsePoemFromOnePageOfCollection 获取诗集中的单首诗歌，返回只有一首诗歌的诗歌数组
// 例子页面：http://www.shiku.org/shiku/xs/haizi/100.htm
func (t *XianDaiShi) parsePoemFromOnePageOfCollection() []models.Poem {
	gbkTitle := t.doc.Find("body").Find("h1").Text()
	titleBytes := []byte(gbkTitle)
	title := strings.TrimSpace(util.GBK2Unicode(titleBytes))

	gbkContent := t.doc.Find("pre").Text()
	poemContentBytes := []byte(gbkContent)
	content := strings.TrimSpace(util.GBK2Unicode(poemContentBytes))

	if title == "诗人简介" {
		return t.Poems
	}

	poem := models.Poem{
		AuthorID: t.Poet.ID,
		Author:   t.Poet.Name,
		Source:   t.doc.Url.String(),
		Title:    title,
		Content:  content,
	}

	t.Poems = append(t.Poems, poem)

	return t.Poems
}

// parsePoems 通过在标题前加分隔符后解析纯文本的方式解析页面
func (t *XianDaiShi) parsePoems() []models.Poem {
	titles := make([]string, 0, 0)

	has999999 := false
	t.doc.Find("body").Find("a").Each(func(i int, s *goquery.Selection) {
		href, existHref := s.Attr("href")
		if existHref {
			if strings.Contains(href, "#") {
				gbkTitle := s.Text()
				titleBytes := []byte(gbkTitle)
				title := strings.TrimSpace(util.GBK2Unicode(titleBytes))
				title = strings.Replace(title, ALinkTitleSep, "", -1)
				titles = append(titles, title)
			}
		}

		name, existName := s.Attr("name")

		if existName {
			if strings.Contains("0123456789", name[0:1]) {
				t.doc.Find("body").Find("a[name=\"" + name + "\"]").AppendHtml(ContentTitleSep)
			}
		}

		if name == "999999" {
			has999999 = true
		}
	})

	gbkText := t.doc.Text()
	TextBytes := []byte(gbkText)
	text := strings.TrimSpace(util.GBK2Unicode(TextBytes))
	// 去掉页脚文字：中国诗歌库 中华诗库 中国诗典 中国诗人 中国诗坛 首页
	text = strings.Replace(text, "中国诗歌库", "", -1)
	text = strings.Replace(text, "中华诗库", "", -1)
	text = strings.Replace(text, "中国诗典", "", -1)
	text = strings.Replace(text, "中国诗人", "", -1)
	text = strings.Replace(text, "中国诗坛", "", -1)
	text = strings.Replace(text, "首页", "", -1)
	// 可能页面底部有以_uacct开头的js文字
	// 例如页面： http://www.shiku.org/shiku/ws/wg/corneille.htm
	index := strings.Index(text, "_uacct")
	if index > 0 {
		text = text[0:index]
	}

	text = strings.TrimSpace(text)
	textArr := strings.Split(text, ContentTitleSep)
	if strings.Contains(text, ContentTitleSep) {
		content := textArr[1:]
		if has999999 {
			content = textArr[1 : len(textArr)-1]
		}

		fmt.Println("解析到的诗歌体数量为：", len(content))
		fmt.Println("解析到的诗歌标题数量为：", len(titles))

		// 标题非链接的情况，获取不到标题，例如： http://www.shiku.org/shiku/ws/wg/corneille.htm
		// 标题链接少于实际的诗歌体数量的情况，例如：http://www.shiku.org/shiku/ws/wg/mallarme.htm
		if len(titles) != len(content) {
			for _, whole := range content {
				whole = strings.TrimLeft(whole, " ")
				title := strings.Split(whole, " ")[0]
				str := strings.TrimSpace(whole)
				content := strings.TrimLeft(str, title)
				content = strings.Replace(content, ContentTitleSep, "", -1)

				poem := models.Poem{
					AuthorID: t.Poet.ID,
					Author:   t.Poet.Name,
					Source:   t.doc.Url.String(),
					Title:    title,
					Content:  content,
				}

				t.Poems = append(t.Poems, poem)
			}
		} else {
			count := len(content)
			for i := 0; i < count; i++ {
				whole := strings.TrimSpace(ContentTitleSep + content[i])
				title := titles[i]
				content := strings.Replace(whole, ContentTitleSep+title, "", -1)
				content = strings.Replace(content, ContentTitleSep, "", -1)

				// 网页本身有错误，标题为空
				// 标题与内容混在一起：http://www.shiku.org/shiku/xs/hanzuorong.htm
				if title == "" {
					content = strings.Trim(content, " ")
					content = strings.Trim(content, "\n")
					arr := strings.Split(content, "\n")
					if len(arr) > 1 {
						title = strings.Replace(arr[0], " ", "", -1)
					}
				}

				poem := models.Poem{
					AuthorID: t.Poet.ID,
					Author:   t.Poet.Name,
					Source:   t.doc.Url.String(),
					Title:    title,
					Content:  content,
				}

				t.Poems = append(t.Poems, poem)
			}
		}
	}

	return t.Poems
}
