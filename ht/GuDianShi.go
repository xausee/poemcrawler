package ht

import (
	"PoemCrawler/models"
	"PoemCrawler/util"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/gocrawl"
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/mgo.v2/bson"
	"log"
	"regexp"
)

// 处理古代诗歌的类型
// 页面样例 http://www.shiku.org/shiku/gs/beichao.htm
type GuDianShi struct {
	uctx    *gocrawl.URLContext
	res     *http.Response
	doc     *goquery.Document
	charset string // 页面编码
	Poet    *models.Poet
}

func NewGuDianShi(uctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) *GuDianShi {
	ch := GetCharset(doc)
	poet := GetPoet(uctx, doc, ch)
	poet.ID = bson.NewObjectId().Hex()
	//poet := GetPoet(uctx, doc, "gbk")
	return &GuDianShi{
		uctx:    uctx,
		res:     res,
		doc:     doc,
		charset: ch,
		Poet:    &poet,
	}
}

// http://www.shiku.org/shiku/gs/shijing.htm
func (t GuDianShi) GetShiJing() (poems []models.Poem) {
	gbkText := t.doc.Text()
	textBytes := []byte(gbkText)
	text := strings.TrimSpace(util.GBK2Unicode(textBytes))
	text = util.RemoveBottomText(text)
	text = strings.Replace(text, "选自国学站点", "", -1)
	text = strings.TrimSpace(text)

	d := "○"

	arr := strings.Split(text, d)
	if len(arr) < 2 {
		return
	}

	arr = arr[1:]

	for _, p := range arr {
		p := strings.TrimSpace(p)
		i := strings.Index(p, "\n")
		// p = "○南陔（今佚）" 这种情况
		if i == -1 {
			continue
		}
		title := p[0:i]
		content := p[i:]

		poem := models.Poem{
			AuthorID: t.Poet.ID,
			Author:   t.Poet.Name,
			Source:   t.uctx.URL().String(),
			Title:    title,
			Volume:   "",
			Content:  content,
		}
		poems = append(poems, poem)
	}

	return
}

// http://www.shiku.org/shiku/gs/yuefu/yfsj_038.htm
// http://www.shiku.org/shiku/gs/yuefu/yfsj_069.htm
func (t GuDianShi) GetYueFuTongQian() (poems []models.Poem) {
	gbkText := t.doc.Text()
	textBytes := []byte(gbkText)
	text := strings.TrimSpace(util.GBK2Unicode(textBytes))
	text = util.RemoveBottomText(text)
	text = strings.TrimSpace(text)

	// 标题样式：【仙吕】一半儿　四景
	d := "【"
	f := "】"

	arr := strings.Split(text, d)
	if len(arr) < 2 {
		return
	}

	arr = arr[1:]

	n := ""  //作者名字
	tl := "" // 标题
	y1 := "【同前】"
	y2 := "【同上】"
	for _, p := range arr {
		j := strings.Index(p, f)
		k := strings.Index(p, "\n")
		if j != -1 && k != -1 && k >= j {
			if !strings.Contains(p[0:j], "同前") && !strings.Contains(p[0:j], "同上") {
				tl = p[0:j]
			}

			r := p[j:k]
			r = strings.Replace(r, f, "", -1)
			r = strings.Replace(r, " ", "", -1)

			if r != "" {
				n = r
			} else {
				// 遇到标题说明时解析完后不记入诗歌
				l := d + p[0:k]
				if l != y1 && l != y2 {
					continue
				}
			}
		}

		p := strings.TrimSpace(p)
		i := strings.Index(p, "\n")
		if i == -1 {
			continue
		}
		title := d + p[0:i]
		title = strings.TrimSpace(title)

		content := p[i:]
		poem := models.Poem{
			AuthorID: "",
			Author:   n,
			Source:   t.uctx.URL().String(),
			Title:    tl,
			Volume:   "",
			Content:  content,
		}
		poems = append(poems, poem)
	}

	return poems
}

// http://www.shiku.org/shiku/gs/yuefu/yfsj_100.htm
func (t GuDianShi) GetYueFu1() (poems []models.Poem) {
	gbkText := t.doc.Text()
	textBytes := []byte(gbkText)
	text := strings.TrimSpace(util.GBK2Unicode(textBytes))
	text = util.RemoveBottomText(text)
	text = strings.TrimSpace(text)

	// 标题样式：【仙吕】一半儿　四景
	d := "【"
	f := "】"

	arr := strings.Split(text, d)
	if len(arr) < 2 {
		return
	}

	arr = arr[1:]

	n := "" // 诗人名字
	for _, p := range arr {
		j := strings.Index(p, f)
		k := strings.Index(p, "\n")

		if j != -1 && k != -1 && k >= j {
			r := p[j:k]
			r = strings.Replace(r, f, "", -1)
			r = strings.Replace(r, " ", "", -1)
			if r != "" {
				n = r
				t.Poet.Name = n
			}
		}

		p := strings.TrimSpace(p)
		i := strings.Index(p, "\n")
		if i == -1 {
			continue
		}
		title := d + p[0:i]
		title = strings.Replace(title, d, "", -1)
		title = strings.Replace(title, f, "", -1)
		title = strings.TrimSpace(title)

		content := p[i:]

		poem := models.Poem{
			AuthorID: t.Poet.ID,
			Author:   t.Poet.Name,
			Source:   t.uctx.URL().String(),
			Title:    title,
			Volume:   "",
			Content:  content,
		}
		poems = append(poems, poem)
	}

	return
}

func (t GuDianShi) GetYueFu() (poems []models.Poem) {
	gbkText := t.doc.Text()
	textBytes := []byte(gbkText)
	text := strings.TrimSpace(util.GBK2Unicode(textBytes))
	f1 := "【同前】"
	f2 := "【同上】"
	if strings.Contains(text, f1) || strings.Contains(text, f2) {
		poems = t.GetYueFuTongQian()
	} else {
		poems = t.GetYueFu1()
	}

	return
}

func (t GuDianShi) GetYuanQu() (poems []models.Poem) {
	gbkText := t.doc.Find("title").Text()
	bytes := []byte(gbkText)
	title := util.GBK2Unicode(bytes)
	log.Println(title)

	if strings.Contains(title, "全元散曲") {
		poems = t.GetQuanYuanSanQu()
	} else if strings.Contains(title, "全元曲") {
		poems = t.GetQuanYuanZaJu()
	}

	return
}

func (t GuDianShi) GetQuanYuanZaJu() (poems []models.Poem) {
	// TODO http://www.shiku.org/shiku/gs/qyq/ 页面上的杂居解析

	return
}

// http://www.shiku.org/shiku/gs/qysq/s018.htm
// TODO http://www.shiku.org/shiku/gs/qysq/s072.htm
func (t GuDianShi) GetQuanYuanSanQu() (poems []models.Poem) {
	t.Poet.Intro = GetPoetIntro(t.doc, t.charset)
	gbkText := t.doc.Text()
	textBytes := []byte(gbkText)
	text := strings.TrimSpace(util.GBK2Unicode(textBytes))
	text = util.RemoveBottomText(text)
	text = strings.TrimSpace(text)

	// 标题样式：【仙吕】一半儿　四景
	d := "【"
	f := "】"

	arr := strings.Split(text, d)
	if len(arr) < 2 {
		return
	}

	arr = arr[1:]

	for _, p := range arr {
		p := strings.TrimSpace(p)
		i := strings.Index(p, "\n")
		if i == -1 {
			continue
		}
		title := p[0:i]
		title = strings.TrimSpace(title)
		title = strings.Replace(title, d, "", -1)
		title = strings.Replace(title, f, ".", -1)
		content := p[i:]
		content = strings.Replace(content, d, "", -1)
		content = strings.Replace(content, f, "", -1)

		l := len(title)
		if l > 90 {
			j := strings.Index(p, f)
			title = p[:j]
			content = p[j:]
			content = strings.Replace(content, d, "", -1)
			content = strings.Replace(content, f, "", -1)
		}

		poem := models.Poem{
			AuthorID: t.Poet.ID,
			Author:   t.Poet.Name,
			Source:   t.uctx.URL().String(),
			Title:    title,
			Volume:   "",
			Content:  content,
		}
		poems = append(poems, poem)
	}

	return
}

// 获取全唐诗 http://www.shiku.org/shiku/gs/tangshi/qts_0487.htm
func (t GuDianShi) GetTangShi() (poems []models.Poem) {
	poems = make([]models.Poem, 0, 0)

	gbkText := t.doc.Text()
	textBytes := []byte(gbkText)
	text := strings.TrimSpace(util.GBK2Unicode(textBytes))
	text = util.RemoveBottomText(text)
	text = strings.TrimSpace(text)
	lines := strings.Split(text, "\n")

	title := ""
	author := ""
	content := ""
	Volume := ""
	for _, line := range lines {
		line = strings.Replace(line, " ", "", -1)
		if line == "" {
			continue
		}

		isTitle := false
		if strings.Contains(line, "「") && strings.Contains(line, "」") {
			if title != "" {
				poem := models.Poem{
					AuthorID: t.Poet.ID,
					Author:   author,
					Source:   t.uctx.URL().String(),
					Title:    title,
					Volume:   Volume,
					Content:  content,
				}
				poems = append(poems, poem)
				title = ""
			}

			Volume = line[0:strings.Index(line, "「")]
			Volume = strings.Replace(Volume, "_", "-", -1)
			title = line[strings.Index(line, "「"):strings.Index(line, "」")]
			title = strings.Replace(title, "「", "", -1)
			author = line[strings.Index(line, "」"):]
			author = strings.Replace(author, "」", "", -1)
			content = ""
			isTitle = true
		}

		if !isTitle {
			content += line + "\r\n"
		}
	}
	return
}

// 获取全宋词 http://www.shiku.org/shiku/gs/songci/0162.htm
func (t GuDianShi) GetSongCi() (poems []models.Poem) {
	poems = make([]models.Poem, 0, 0)

	t.doc.Find("p[align=\"center\"]").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			gbkText := s.Text()
			textBytes := []byte(gbkText)
			text := util.GBK2Unicode(textBytes)
			name := strings.TrimSpace(text)
			t.Poet.Name = name
		} else {
			gbkText := s.Text()
			textBytes := []byte(gbkText)
			text := util.GBK2Unicode(textBytes)
			title := strings.TrimSpace(text)
			s.Next().Text()

			gbkText = s.Next().Text()
			textBytes = []byte(gbkText)
			text = util.GBK2Unicode(textBytes)
			text = strings.TrimLeft(text, "\r")
			text = strings.TrimLeft(text, "\n")
			text = strings.TrimRight(text, "\r")
			text = strings.TrimRight(text, "\n")
			content := text
			poem := models.Poem{
				AuthorID: t.Poet.ID,
				Author:   t.Poet.Name,
				Source:   t.uctx.URL().String(),
				Title:    title,
				Content:  content,
			}
			poems = append(poems, poem)
		}
	})
	return
}

// 获取唐诗三百首 http://www.shiku.org/shiku/gs/tangdai.htm
func (t GuDianShi) GetTangShiSanBaiShou() (poems []models.Poem) {
	poems = make([]models.Poem, 0, 0)

	gbkcontent := t.doc.Find("a[name=\"001\"]").Next().Text()
	contentBytes := []byte(gbkcontent)
	content := strings.TrimSpace(util.GBK2Unicode(contentBytes))
	poem := models.Poem{
		AuthorID: t.Poet.ID,
		Author:   "张九龄",
		Source:   t.uctx.URL().String(),
		Title:    "感遇四首之一",
		Content:  content,
	}
	poems = append(poems, poem)

	t.doc.Find("content").Find("p").Find("a").Each(func(i int, s *goquery.Selection) {
		_, exist := s.Attr("name")
		if exist {
			gbkTitleLine := s.Parent().Text()
			titleLineBytes := []byte(gbkTitleLine)
			titleLine := strings.TrimSpace(util.GBK2Unicode(titleLineBytes))
			titleLine = strings.Replace(titleLine, "+", "", -1)

			author := strings.Split(titleLine, "：")[0][3:]
			title := strings.Split(titleLine, "：")[1]
			title = strings.Replace(title, " ", "", -1)

			gbkcontent := s.Next().Text()
			contentBytes := []byte(gbkcontent)
			content := strings.TrimSpace(util.GBK2Unicode(contentBytes))
			content = strings.Replace(title, " ", "", -1)

			poem := models.Poem{
				AuthorID: t.Poet.ID,
				Author:   author,
				Source:   t.uctx.URL().String(),
				Title:    title,
				Content:  content,
			}
			poems = append(poems, poem)
		}
	})
	return
}

// 获取没有诗歌名链接的页面诗歌，如：http://www.shiku.org/shiku/ws/ww/blake.htm 和 http://www.shiku.org/shiku/ws/ww/horace.htm
func (t GuDianShi) GetPoemsWithoutTitleLink(selector string) (poems []models.Poem) {
	poet := t.Poet
	poems = make([]models.Poem, 0, 0)

	t.doc.Find("body").Find(selector).Each(func(i int, s *goquery.Selection) {
		text := util.TrimBottomAndToCharset(s.Text(), t.charset)

		// 去掉额外的符号
		text = strings.Replace(text, "聽", "", -1)
		text = strings.TrimSpace(text)

		n := strings.Index(text, "\n")
		if n > 0 {
			title := text[0:n]
			content := text[n:]
			poem := models.Poem{
				AuthorID: poet.ID,
				Author:   poet.Name,
				Source:   t.uctx.URL().String(),
				Title:    title,
				Content:  content,
			}
			poems = append(poems, poem)
		}
	})
	return
}

// 通过在标题前加分隔符后解析纯文本的方式解析页面
func (t GuDianShi) GetPoems() (poems []models.Poem) {
	links := len(t.doc.Find("body").Find("a").Nodes)
	// 页面上只有页脚的6个超链接，说明此页面没有诗歌标题链接，则调用GetPoemsWithoutTitleLink()解析诗歌
	if links == 6 {
		selector := "p[align=\"left\"]"
		poems = t.GetPoemsWithoutTitleLink(selector)
		if len(poems) == 0 {
			selector = "p"
			poems = t.GetPoemsWithoutTitleLink(selector)
		}
		return
	}
	log.Println(t.charset)

	titles := make([]string, 0, 0)
	poet := t.Poet
	poems = make([]models.Poem, 0, 0)

	has999999 := false
	isRomaId := false
	t.doc.Find("body").Find("a").Each(func(i int, s *goquery.Selection) {
		href, existHref := s.Attr("href")
		if existHref {
			if strings.Contains(href, "#") {
				gbkTitle := s.Text()
				titleBytes := []byte(gbkTitle)
				title := util.ToUnicode(titleBytes, t.charset)
				title = strings.TrimSpace(title)
				title = strings.Replace(title, ALinkTitleSep, "", -1)
				title = strings.Replace(title, "\n", "", -1)
				title = strings.Replace(title, "聽", "", -1)
				titles = append(titles, title)
			}
		}

		name, existName := s.Attr("name")

		if existName {
			if strings.Contains("0123456789", name[0:1]) {
				t.doc.Find("body").Find("a[name=\"" + name + "\"]").AppendHtml(ContentTitleSep)
			}

			id, existId := s.Attr("id")
			if existId && id == name {
				c := []byte(name)
				if check, _ := regexp.Match("I{0,3}", c); check == true {
					//if check, _ := regexp.Match("(?:M{0,3})(?:D?C{0,3}|C[DM])(?:L?X{0,3}|X[LC])(?:V?I{0,3}|I[VX])", c); check == true {
					isRomaId = true
					s.AppendHtml(ContentTitleSep)
					//t.doc.Find("body").Find("a[name=\"" + name + "\"]").AppendHtml(ContentTitleSep)
				}
			}
		}

		if name == "999999" {
			has999999 = true
		}
	})

	text := util.TrimBottomAndToCharset(t.doc.Text(), t.charset)
	text = strings.Replace(text, "聽", "", -1)

	textArr := strings.Split(text, ContentTitleSep)
	if strings.Contains(text, ContentTitleSep) {
		textContent := textArr[1:]
		if has999999 {
			textContent = textArr[1 : len(textArr)-1]
		}

		log.Println("解析到的诗歌体数量为：", len(textContent))
		log.Println("解析到的诗歌标题数量为：", len(titles))

		// 标题非链接的情况，获取不到标题，例如： http://www.shiku.org/shiku/ws/wg/corneille.htm
		if len(titles) != len(textContent) {
			for _, whole := range textContent {
				var title string
				whole = strings.TrimLeft(whole, " ")
				if strings.Contains(whole, ContentTitleSep1) {
					title = strings.Split(whole, ContentTitleSep1)[0]
				} else if isRomaId {
					title = strings.Split(whole, "\n")[0]
				} else {
					title = strings.Split(whole, " ")[0]
				}

				str := strings.TrimSpace(whole)
				content := strings.TrimLeft(str, title)
				content = strings.Replace(content, ContentTitleSep, "", -1)
				content = strings.Replace(content, ContentTitleSep1, "", -1)

				if isRomaId && strings.Replace(content, " ", "", -1) == "" {
					continue
				}

				if isRomaId && strings.Replace(title, " ", "", -1) == "" {
					log.Println(whole)
					continue
				}

				title = strings.Replace(title, "\n", "", -1)
				poem := models.Poem{
					AuthorID: poet.ID,
					Author:   poet.Name,
					Source:   t.uctx.URL().String(),
					Title:    title,
					Content:  content,
				}
				poems = append(poems, poem)
			}
		} else {
			count := len(textContent)
			for i := 0; i < count; i++ {
				whole := strings.TrimSpace(ContentTitleSep + textContent[i])
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
				title = strings.Replace(title, "\n", "", -1)

				poem := models.Poem{
					AuthorID: poet.ID,
					Author:   poet.Name,
					Source:   t.uctx.URL().String(),
					Title:    title,
					Content:  content,
				}

				poems = append(poems, poem)
			}
		}
	}

	return
}
