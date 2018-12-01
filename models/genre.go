package models

type PoetAddress struct {
	Name       string // 诗人名
	UrlAddress string // url地址
}

type Genre struct {
	ID                  string        // ID号
	Name                string        // 流派名
	N                   int           // 自增序号
	Description         string        // 简介
	Chronology          string        // 所属年代
	PoetAddresses       []PoetAddress // 诗人网页地址（最后的路径）
	TimeStamp           string        // 创建时间戳
	LastUpdateTimeStamp string        // 最后更新时间戳
}
