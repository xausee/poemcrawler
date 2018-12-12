package db

import (
	"PoemCrawler/models"
	"PoemCrawler/util"
)

func Save(hasPoet bool, isPoemCollection bool, poet models.Poet, poemType string, poems []models.Poem, url string) {
	msg := ""
	if hasPoet {
		err := util.CheckPoet(poet)
		if err != nil {
			msg = err.Error()
		} else {
			if !isPoemCollection {
				SavePoet(poet)
			}
		}
	}

	err := util.CheckPoems(poems)
	count := len(poems)

	if err != nil {
		msg = err.Error()
		if msg == "诗歌标题过长，解析可能有误" {
			SavePoems(poems, poemType)
		} else {
			count = 0
		}
	} else {
		SavePoems(poems, poemType)
	}

	SaveAddress(url, msg, count)
}

