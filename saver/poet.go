package db

import (
	"PoemCrawler/models"
	"github.com/mozillazg/go-pinyin"
	"log"
	"strings"
	"time"
	"unicode"
)

// SavePoet 保存诗人信息
func SavePoet(p models.Poet) {
	db, err := models.NewDBManager()
	if err != nil {
		log.Println(err.Error())
	}
	defer db.Close()

	c := db.Session.DB(models.CONFIG.Mongo.DB).C(models.PoetCollection)

	r := []rune(p.Name)

	digits := map[string]string{
		"0": "L",
		"1": "Y",
		"2": "E",
		"3": "S",
		"4": "S",
		"5": "W",
		"6": "L",
		"7": "Q",
		"8": "B",
		"9": "J",
	}

	//if unicode.IsLetter(r[0]) {
	// // unicode.IsLetter 方法无法判断汉字
	//	// 字母开头的情况
	//	p.AlphabetIndex = string(r[0])
	//}
	letters:="ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if strings.Contains(string(r[0]), letters){
		// 字母开头的情况
		p.AlphabetIndex = string(r[0])
	}else if unicode.IsDigit(r[0]) {
		// 阿拉伯数字的情况
		p.AlphabetIndex = digits[string(r[0])]
	} else {
		// 汉字的情况
		p.AlphabetIndex = strings.ToUpper(pinyin.LazyPinyin(p.Name, pinyin.NewArgs())[0])[0:1]
	}

	p.Avatar = ""
	p.TimeStamp = time.Now().Format("2006-01-02 15:04:05")
	p.LastUpdateTimeStamp = time.Now().Format("2006-01-02 15:04:05")

	err = c.Insert(p)
	if err != nil {
		log.Println("保存诗人信息失败：", p.Name)
	}
	log.Println("保存诗人信息成功：", p.Name)
}
