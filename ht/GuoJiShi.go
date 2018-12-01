package ht

import (
	"PoemCrawler/models"
	"PoemCrawler/util"
	"log"
	"net/http"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"github.com/PuerkitoBio/gocrawl"
	"github.com/PuerkitoBio/goquery"
	"regexp"
)

const (
	ALinkTitleSep    = "++++++++++++++++++++++++"
	ContentTitleSep  = "========================"
	ContentTitleSep1 = "111111111111111111111111"
)

// 处理国际诗歌的类型
// 首页 http://www.shiku.org/shiku/ws/wg/index.htm
type GuoJiShi struct {
	uctx    *gocrawl.URLContext
	res     *http.Response
	charset string // 页面编码
	doc     *goquery.Document
	Poet    *models.Poet
}

func NewGuoJiShi(uctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) *GuoJiShi {
	ch := GetCharset(doc)
	poet := GetPoet(uctx, doc, ch)
	poet.ID = bson.NewObjectId().Hex()

	return &GuoJiShi{
		uctx:    uctx,
		res:     res,
		doc:     doc,
		charset: ch,
		Poet:    &poet,
	}
}

func GetCharset(doc *goquery.Document) string {
	charset := "gbk"
	t, e := doc.Find("head").Html()
	if e != nil {
		log.Println(e)
		return charset
	}

	r := regexp.MustCompile(`charset=.*"/`)
	a := r.FindAllString(t, -1)
	if len(a) > 0 {
		charset = a[0]
		charset = strings.Replace(charset, "charset=", "", -1)
		charset = strings.Replace(charset, "\"", "", -1)
		charset = strings.Replace(charset, "/", "", -1)
	}

	if strings.Contains(charset, "gb") || strings.Contains(charset, "GB") {
		log.Println("原始编码为" + charset + " 改为gbk编码")
		return "gbk"
	}

	return charset
}

func GetFirstPoemTitleWithSep(doc *goquery.Document, charset string) string {
	titles := make([]string, 0, 0)
	var title string
	doc.Find("body").Find("a").Each(func(i int, s *goquery.Selection) {
		href, existHref := s.Attr("href")
		if existHref {
			if strings.HasPrefix(href, "#") {
				gbkTitle := s.Text()
				titleBytes := []byte(gbkTitle)
				title = util.ToUnicode(titleBytes, charset)
				title = strings.TrimSpace(title)
				titles = append(titles, title+ALinkTitleSep)
				s.AppendHtml(ALinkTitleSep)
				return
			}
		}
	})
	if len(titles) > 0 {
		return titles[0]
	}

	return ""
}

func GetFirstPoemTitleWithSep1FromPoemBody(doc *goquery.Document, charset string) string {
	titles := make([]string, 0, 0)
	var title string
	doc.Find("body").Find("h2").Each(func(i int, s *goquery.Selection) {
		gbkTitle := s.Text()
		titleBytes := []byte(gbkTitle)
		title = util.ToUnicode(titleBytes, charset)
		title = strings.TrimSpace(title)
		titles = append(titles, title+ContentTitleSep1)
		s.AppendHtml(ContentTitleSep1)
	})
	if len(titles) > 0 {
		return titles[0]
	}

	return ""
}

func GetPoetIntro(doc *goquery.Document, charset string) string {
	intro := ""

	doc.Find("tbody").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			gbkStr := s.Text()
			bytes := []byte(gbkStr)
			intro = util.ToUnicode(bytes, charset)
			intro = strings.TrimSpace(intro)
			intro = strings.Replace(intro, "\n", "", -1)
		}
		return
	})

	return intro
}

// http://www.shiku.org/shiku/gs/chuci.htm
func GetPoet(uctx *gocrawl.URLContext, doc *goquery.Document, charset string) (poet models.Poet) {
	if strings.Contains(doc.Url.Path, "chuci.htm") {
		poet = models.Poet{
			Name:   "屈原",
			Intro:  "",
			Source: uctx.URL().String(),
		}
		return
	}

	gbkStr := doc.Find("title").Text()
	bytes := []byte(gbkStr)
	title := util.ToUnicode(bytes, charset)
	title = strings.TrimSpace(title)

	var name string
	arr := strings.Split(title, "::")

	for _, v := range arr {
		if strings.Contains(v, "诗选") || strings.Contains(v, "诗集") || strings.Contains(v, "全集") {
			name = strings.TrimSpace(strings.Split(v, "诗全集")[0])
			name = strings.TrimSpace(strings.Split(name, "诗选")[0])
			name = strings.TrimSpace(strings.Split(name, "诗集")[0])
			name = strings.TrimSpace(strings.Split(name, "全集")[0])
		}
	}

	if name == "" {
		name = GetPoetNameFrom(doc, "h1", charset)
	}

	if strings.Contains(name, "节选") {
		name = GetPoetNameFrom(doc, "p", charset)
	}

	var intro string
	ft := GetFirstPoemTitleWithSep(doc, charset)
	if ft == "" {
		// 处理页面上诗人简介下面的标题不是链接的情况
		// http://www.shiku.org/shiku/ws/wg/corneille.htm
		ft = GetFirstPoemTitleWithSep1FromPoemBody(doc, charset)
	}

	if ft != "" {
		gbkStr = doc.Find("body").Text()
		bytes = []byte(gbkStr)
		text := util.ToUnicode(bytes, charset)
		text = strings.TrimSpace(text)

		index := strings.Index(text, ft)
		if index > 0 {
			text = text[0:index]
		}

		intro = strings.TrimSpace(text)
		intro = strings.Replace(intro, name+"诗选", "", 1)
		intro = strings.Replace(intro, name+"诗集", "", 1)
		intro = strings.Replace(intro, name+"全集", "", 1)
		intro = strings.Replace(intro, name+"诗全集", "", 1)
		intro = strings.Replace(intro, "\r", "", -1)
		intro = strings.Replace(intro, "\n", "", -1)
	}

	poet = models.Poet{
		Name:   name,
		Intro:  intro,
		Source: uctx.URL().String(),
	}

	return poet
}

// 获取没有诗歌名链接的页面诗歌，如：http://www.shiku.org/shiku/ws/ww/blake.htm 和 http://www.shiku.org/shiku/ws/ww/horace.htm
func (t GuoJiShi) GetPoemsWithoutTitleLink(selector string) (poems []models.Poem) {
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

// http://www.shiku.org/shiku/ws/wg/yiger.htm
// 作者为“xxx的史诗”的类型页面
func (t GuoJiShi) GetPoemsOfShiShi() (poems []models.Poem) {
	poems = make([]models.Poem, 0, 0)
	title := GetPoetNameFrom(t.doc, "h1", t.charset)
	text := util.TrimBottomAndToCharset(t.doc.Text(), t.charset)
	i := strings.Index(text, "的史诗")

	var content string
	if i != -1 {
		content = text[i+len("的史诗"):]
	}

	poem := models.Poem{
		AuthorID: t.Poet.ID,
		Author:   t.Poet.Name,
		Source:   t.uctx.URL().String(),
		Title:    title,
		Content:  content,
	}
	poems = append(poems, poem)

	return
}

// http://www.shiku.org/shiku/ws/zg/shijing.htm
func (t GuoJiShi) GetPoemsWsZgShiJing() (poems []models.Poem) {
	poet := t.Poet
	poet.Name = "诗经 " + poet.Name
	poems = make([]models.Poem, 0, 0)

	text := util.TrimBottomAndToCharset(t.doc.Text(), t.charset)

	delimiter := "*************************"
	reg := regexp.MustCompile(`(\d+)\.`)
	text = reg.ReplaceAllString(text, "${n}"+delimiter)
	a := strings.Split(text, delimiter)
	for _, e := range a {
		i := strings.Index(e, "\n")
		title := e[0:i]
		title = strings.TrimSpace(title)
		content := e[i:]
		content = strings.TrimSpace(content)

		poem := models.Poem{
			AuthorID: poet.ID,
			Author:   poet.Name,
			Source:   t.uctx.URL().String(),
			Title:    title,
			Content:  content,
		}
		poems = append(poems, poem)
	}

	return
}

// http://www.shiku.org/shiku/ws/ww/apollinaire.htm
func (t GuoJiShi) GetPoemsApollinaire() (poems []models.Poem) {
	poet := t.Poet
	text := util.TrimBottomAndToCharset(t.doc.Text(), t.charset)

	t1 := "Mai"
	t2 := "NUIT RHÉNANE"
	t2t := "NUIT RH"
	t3 := "LA LORELEY"

	i1 := strings.Index(text, t1)
	i2 := strings.Index(text, t2t)
	i3 := strings.Index(text, t3)

	poem1 := models.Poem{
		AuthorID: poet.ID,
		Author:   poet.Name,
		Source:   t.uctx.URL().String(),
		Title:    t1,
		Content:  text[i1+len(t1) : i2],
	}

	poem2 := models.Poem{
		AuthorID: poet.ID,
		Author:   poet.Name,
		Source:   t.uctx.URL().String(),
		Title:    t2,
		Content:  text[i2+len(t2) : i3],
	}

	poem3 := models.Poem{
		AuthorID: poet.ID,
		Author:   poet.Name,
		Source:   t.uctx.URL().String(),
		Title:    t3,
		Content:  text[i3+len(t3):],
	}
	poems = append(poems, poem1)
	poems = append(poems, poem2)
	poems = append(poems, poem3)

	return
}

// 通过在标题前加分隔符后解析纯文本的方式解析页面
func (t GuoJiShi) GetPoems() (poems []models.Poem) {
	links := len(t.doc.Find("body").Find("a").Nodes)
	// 页面上只有页脚的6个超链接，说明此页面没有诗歌标题链接，则调用GetPoemsWithoutTitleLink()解析诗歌
	if links == 6 {
		text := util.TrimBottomAndToCharset(t.doc.Text(), t.charset)
		if strings.Contains(text, "的史诗") {
			return t.GetPoemsOfShiShi()
		}

		selector := "p[align=\"left\"]"
		poems = t.GetPoemsWithoutTitleLink(selector)
		if len(poems) == 0 {
			selector = "p"
			poems = t.GetPoemsWithoutTitleLink(selector)
		}

		return
	}

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
			//romaNumRexp := "^M{0,4}(CM|CD|D?C{0,3})(XC|XL|L?X{0,3})(IX|IV|V?I{0,3})"
			//if isRoma, _:= regexp.MatchString(romaNumRexp, name) ; isRoma == true{
			//	fmt.Println(name)
			//	t.doc.Find("body").Find("a[name=\"" + name + "\"]").AppendHtml(ContentTitleSep)
			//}
		}

		if name == "999999" {
			has999999 = true
		}
	})

	text := util.TrimBottomAndToCharset(t.doc.Text(), t.charset)

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

// GetDanDingShenQu 获取但丁的神曲
// http://www.shiku.org/shiku/ws/wg/dante/index.htm
func (t GuoJiShi) GetDanDingShenQu() (poems []models.Poem) {
	poems = make([]models.Poem, 0, 0)

	gbkTitle := t.doc.Find("body").Find("p[align=\"center\"]").Text()
	bytes := []byte(gbkTitle)
	order := util.ToUnicode(bytes, t.charset)
	order = strings.TrimSpace(order)

	sep := "**************"
	t.doc.Find("body").Find("p[align=\"center\"]").AppendHtml(sep)

	gbkText := t.doc.Text()
	bytes = []byte(gbkText)
	text := util.ToUnicode(bytes, t.charset)
	text = strings.TrimSpace(text)

	col := strings.Split(text, order)[0]
	col = strings.Replace(col, "[意] 但丁：", "", -1)
	col = strings.Replace(col, " ", "", -1)
	col = strings.Replace(col, "\r", "", -1)
	col = strings.Replace(col, "\n", "", -1)
	title := col + "·" + order

	text = strings.Split(text, sep)[1]
	content := util.RemoveBottomText(text)
	content = strings.TrimSpace(content)

	poet := &models.Poet{
		ID:   bson.NewObjectId().Hex(),
		Name: "但丁",
	}

	t.Poet = poet

	poem := models.Poem{
		AuthorID: t.Poet.ID,
		Author:   t.Poet.Name,
		Source:   t.uctx.URL().String(),
		Title:    title,
		Content:  content,
	}

	poems = append(poems, poem)

	return poems
}

// http://www.shiku.org/shiku/ws/zg/tang.htm
func (t GuoJiShi) GetTangShiSanBaiShou() (poems []models.Poem) {
	t.doc.Find("p").Each(func(i int, s *goquery.Selection) {
		id := s.Text()
		if check, _ := regexp.Match("^[0-9]{3}$", []byte(id)); check == true {
			authorName := s.Next().Next().Text()
			title := s.Next().Next().Next().Text()
			content := s.Next().Next().Next().Next().Text()
			source := t.uctx.URL().String()

			if id == "099" {
				authorName = "Cen Can"
				title = "A MESSAGE TO CENSOR Du Fu AT HIS OFFICE IN THE LEFT COURT"
				content = s.Next().Next().Text()
			}
			if id == "168" {
				authorName = "Wei Zhuang"
				title = "A NIGHT THOUGHT ON TERRACE TOWER"
				content = s.Next().Next().Text()
			}
			if id == "169" {
				authorName = "Seng Jiaoran"
				title = "NOT FINDING LU HONGXIAN AT HOME"
				content = s.Next().Next().Text()
			}
			if id == "297" {
				authorName = "Du Mu"
				title = "THE GARDEN OF THE GOLDEN VALLEY"
				content = s.Next().Next().Text()
			}

			poem := models.Poem{
				Author:  authorName,
				Title:   title,
				Content: content,
				Source:  source,
			}
			poems = append(poems, poem)
		}
	})

	return
}

// http://www.shiku.org/shiku/ws/zg/dufu.htm
func (t *GuoJiShi) GetDuFu() (poems []models.Poem) {
	sep := "**************"
	t.doc.Find("hr").Each(func(i int, s *goquery.Selection) {
		s.AppendHtml(sep)
	})

	gbkText := t.doc.Text()
	bytes := []byte(gbkText)
	text := util.ToUnicode(bytes, t.charset)
	text = strings.Replace(text, "聽", "", -1)
	text = util.RemoveBottomText(text)
	index := strings.Index(text, "A VIEW OF TAISHAN")
	text = text[index:]
	text = strings.TrimSpace(text)

	poet := &models.Poet{
		ID:    bson.NewObjectId().Hex(),
		Name:  "Du Fu",
		Intro: t.doc.Find("p[align=\"left\"]").Text(),
	}
	t.Poet = poet

	title := ""
	content := ""
	lines := strings.Split(text, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		line = strings.TrimSpace(line)
		if line == sep {
			poem := models.Poem{
				AuthorID: t.Poet.ID,
				Author:   t.Poet.Name,
				Source:   t.uctx.URL().String(),
				Title:    title,
				Content:  content,
			}
			poems = append(poems, poem)
			title = ""
			content = ""
		}

		lineTmp := strings.Replace(line, "Li Bai", "", -1)
		if check, _ := regexp.Match("[a-z]+", []byte(lineTmp)); check == false {
			if line != sep {
				if title == "" {
					title = line
				} else {
					title += " " + line
				}
			}
		} else {
			content += line + "\n"
		}
	}

	return poems
}

// 获取泰戈尔的诗
// http://www.shiku.org/shiku/ws/wg/tagore2/003.htm
func (t GuoJiShi) GetTaiGeErPoems() (poems []models.Poem) {
	titles := make([]string, 0, 0)
	poems = make([]models.Poem, 0, 0)

	t.doc.Find("body").Find("a").Each(func(i int, s *goquery.Selection) {
		href, existHref := s.Attr("href")
		if existHref {
			if strings.Contains(href, "#") {
				gbkTitle := s.Text()
				bytes := []byte(gbkTitle)
				title := util.ToUnicode(bytes, t.charset)
				title = strings.TrimSpace(title)

				title = strings.Replace(title, ALinkTitleSep, "", -1)
				titles = append(titles, title)
			}
		}

		name, existName := s.Attr("name")
		if existName {
			if strings.Contains("0123456789", name[0:1]) && strings.Replace(s.Text(), " ", "", -1) != "" {
				s.AppendHtml(ContentTitleSep)
			}
		}

	})

	text := util.TrimBottomAndToCharset(t.doc.Text(), t.charset)

	textArr := strings.Split(text, ContentTitleSep)
	if strings.Contains(text, ContentTitleSep) {
		content := textArr[1:]

		log.Println("解析到的诗歌体数量为：", len(content))
		log.Println("解析到的诗歌标题数量为：", len(titles))
		count := len(content)
		for i := 0; i < count; i++ {
			whole := strings.TrimSpace(ContentTitleSep + content[i])
			title := titles[i]
			content := strings.Replace(whole, ContentTitleSep+title, "", -1)
			content = strings.Replace(content, ContentTitleSep, "", -1)
			if title == "" {
				content = strings.Trim(content, " ")
				content = strings.Trim(content, "\n")
				arr := strings.Split(content, "\n")
				if len(arr) > 1 {
					title = strings.Replace(arr[0], " ", "", -1)
				}
			}
			title = strings.Replace(title, " ", "", -1)
			poem := models.Poem{
				Author:  "泰戈尔",
				Source:  t.uctx.URL().String(),
				Title:   title,
				Content: content,
			}

			poems = append(poems, poem)
		}
	}

	return
}

func GetPoetNameFrom(doc *goquery.Document, selector, charset string) (name string) {
	doc.Find("body").Find(selector).Each(func(i int, s *goquery.Selection) {
		gbkStr := s.Text()
		bytes := []byte(gbkStr)
		title := util.ToUnicode(bytes, charset)
		title = strings.TrimSpace(title)

		if i == 0 {
			name = strings.Replace(title, "Poems by", "", -1)
			name = strings.Replace(name, "诗集", "", -1)
			name = strings.Replace(name, "诗选", "", -1)
			name = strings.Replace(name, "全集", "", -1)
			name = strings.Replace(name, "诗全集", "", -1)
			name = strings.TrimSpace(name)
		}

		if name == "" && strings.Contains(title, "的史诗") {
			name = strings.Replace(title, "Poems by", "", -1)
			name = strings.Replace(name, "诗集", "", -1)
			name = strings.Replace(name, "诗选", "", -1)
			name = strings.Replace(name, "全集", "", -1)
			name = strings.Replace(name, "诗全集", "", -1)
			name = strings.TrimSpace(name)
		}
		// http://www.shiku.org/shiku/ws/ww/apollinaire.htm 该页面例外
		name = strings.Replace(name, "Mai", "", -1)
	})
	return
}

func (t GuoJiShi) GetPoemsH2AndP() (poems []models.Poem) {
	poems = make([]models.Poem, 0, 0)

	t.doc.Find("body").Find("h2").Each(func(i int, s *goquery.Selection) {
		gbkFullTitle := s.Text()
		bytes := []byte(gbkFullTitle)
		fullTitle := util.ToUnicode(bytes, t.charset)
		fullTitle = strings.TrimSpace(fullTitle)

		title := strings.TrimSpace(strings.Split(fullTitle, " ")[0])
		subTitle := fullTitle[len(title):]

		// 同一页面有个别节点会多一个<br> 例如 http://www.shiku.org/shiku/ws/wg/transtromo.htm#7
		gbkContent := s.Next().Next().Text()
		if gbkContent == "" {
			gbkContent = s.Next().Next().Next().Text()
		}
		contentBytes := []byte(gbkContent)
		content := util.GBK2Unicode(contentBytes)

		poem := models.Poem{
			Author:   GetPoetNameFrom(t.doc, "h1", t.charset),
			Source:   t.uctx.URL().String(),
			Title:    title,
			SubTitle: subTitle,
			Content:  content,
		}
		poems = append(poems, poem)
	})

	return
}

// http://www.shiku.org/shiku/ws/zg/quyuan.htm
func (t GuoJiShi) GetPoemOfQuYuan() (poems []models.Poem) {
	poems = make([]models.Poem, 0, 0)

	text := util.TrimBottomAndToCharset(t.doc.Text(), t.charset)
	// 去掉额外的符号
	text = strings.Replace(text, "聽", "", -1)
	text = strings.TrimSpace(text)

	p1 := strings.Index(text, "Li Sao (The Lament)") + len("Li Sao (The Lament)")
	p2 := strings.Index(text, "The Fisherman") + len("The Fisherman")
	p3 := strings.Index(text, "CROSSING THE RIVER") + len("CROSSING THE RIVER")

	poem1 := models.Poem{
		Author:   "QuYuan",
		Source:   t.uctx.URL().String(),
		Title:    "Li Sao",
		SubTitle: "",
		Content:  strings.TrimSpace(text[p1:p2]),
	}
	poem2 := models.Poem{
		Author:   "QuYuan",
		Source:   t.uctx.URL().String(),
		Title:    "The Fisherman",
		SubTitle: "",
		Content:  strings.TrimSpace(text[p2:p3]),
	}
	poem3 := models.Poem{
		Author:   "QuYuan",
		Source:   t.uctx.URL().String(),
		Title:    "CROSSING THE RIVER",
		SubTitle: "",
		Content:  strings.TrimSpace(text[p3:]),
	}
	poems = append(poems, poem1)
	poems = append(poems, poem2)
	poems = append(poems, poem3)

	return
}

// http://www.shiku.org/shiku/ws/wg/mandelshtam/005.htm
func (t GuoJiShi) GetSinglePoem() (poems []models.Poem) {
	title := util.ToUnicode([]byte(t.doc.Find("h1").Text()), t.charset)
	title = strings.TrimSpace(title)
	content := util.TrimBottomAndToCharset(t.doc.Text(), t.charset)

	// http://www.shiku.org/shiku/ws/wg/borges/b050.htm
	if len(content) < len(title) {
		delimiter := "======================"
		t.doc.Find("hr").Each(func(i int, s *goquery.Selection) {
			s.AppendHtml(delimiter)
		})

		text := util.ToUnicode([]byte(t.doc.Find("body").Text()), t.charset)
		a := strings.Split(text, delimiter)
		if len(a) < 2 {
			return
		}
		title = a[0]
		title = strings.Replace(title, "\n", "", -1)
		content = a[1]
	} else {
		content = content[len(title):]
	}

	p := strings.TrimLeft(t.doc.Url.Path, "/")
	ps := strings.Split(p, "/")
	n := ps[len(ps)-2]
	t.Poet.Name = n

	poem := models.Poem{
		Author:   t.Poet.Name,
		Source:   t.uctx.URL().String(),
		Title:    title,
		SubTitle: "",
		Content:  content,
	}
	poems = append(poems, poem)
	return poems
}
