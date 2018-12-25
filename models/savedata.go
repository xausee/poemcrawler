package models

type SaveData struct {
	HasPoet          bool   // 是否有诗人数据
	IsPoemCollection bool   // 是否是诗集
	Poet             Poet   // 诗人数据
	PoemType         string //诗歌类型
	Poems            []Poem // 诗歌
	Url              string // 页面地址
}
