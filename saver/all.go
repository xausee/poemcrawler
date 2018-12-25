package db

import (
	"PoemCrawler/models"
	"PoemCrawler/util"
)

// CheckSave 数据校验和保存
func CheckSave(data models.SaveData) {
	msg := ""
	if data.HasPoet {
		err := util.CheckPoet(data.Poet)
		if err != nil {
			msg = err.Error()
		} else {
			if !data.IsPoemCollection {
				SavePoet(data.Poet)
			}
		}
	}

	err := util.CheckPoems(data.Poems)
	count := len(data.Poems)

	if err != nil {
		msg = err.Error()
		if msg == "诗歌标题过长，解析可能有误" {
			SavePoems(data.Poems, data.PoemType)
		} else {
			count = 0
		}
	} else {
		SavePoems(data.Poems, data.PoemType)
	}
	SaveAddress(data.Url, msg, count)
}
