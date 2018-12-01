package util

import (
	"errors"
	"log"
	"strings"

	"github.com/axgle/mahonia"

	"PoemCrawler/models"
)

func GBK2Unicode(data []byte) string {
	var dec mahonia.Decoder

	dec = mahonia.NewDecoder("gbk")

	_, ret, err := dec.Translate(data, true)
	if err != nil {
		log.Println(err)
	}

	return string(ret)
}

func ToUnicode(data []byte, charset string) string {
	var dec mahonia.Decoder

	dec = mahonia.NewDecoder(charset)

	_, ret, err := dec.Translate(data, true)
	if err != nil {
		log.Println(err)
	}

	return string(ret)
}

func TrimBottomAndToCharset(s, ch string) string {
	textBytes := []byte(s)
	text := ToUnicode(textBytes, ch)
	// TODO 拉丁文移除底部链接文字不生效
	text = RemoveBottomText(text)
	text = strings.TrimSpace(text)
	return text
}

// 去掉页脚文字以及以_uacct开头的js文字
func RemoveBottomText(s string) string {
	a := []string{"中国诗歌库", "中华诗库", "中国诗典", "中国诗人", "中国诗坛", "首页"}
	for i := 0; i < len(a); i++ {
		s = strings.Replace(s, a[i], "", -1)
	}

	// 可能页面底部有以_uacct开头的js文字
	// 例如页面： http://www.shiku.org/shiku/ws/wg/corneille.htm
	index := strings.Index(s, "_uacct")
	if index > 0 {
		s = s[0:index]
	}

	return s
}

func CheckPoet(p models.Poet) error {
	name := strings.Replace(p.Name, " ", "", -1)
	//intro := strings.Replace(p.Intro, " ", "", -1)

	if name == "" {
		return errors.New("缺少诗人名字")
	}

	//if intro == "" {
	//	return errors.New("缺少诗人简介")
	//}

	return nil
}

func CheckPoems(ps []models.Poem) error {
	for _, p := range ps {
		author := strings.Replace(p.Author, " ", "", -1)
		title := strings.Replace(p.Title, " ", "", -1)
		body := strings.Replace(p.Content, " ", "", -1)

		if author == "" {
			return errors.New("缺少作者名字")
		}

		if title == "" {
			return errors.New("缺少诗歌标题")
		}

		if len(title) > 20 {
			return errors.New("诗歌标题过长，解析可能有误")
		}

		if body == "" {
			return errors.New("缺少诗歌内容")
		}
	}

	return nil
}
