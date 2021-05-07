package models

type PoetAddress struct {
	Name       string `json:"name"`       // 诗人名
	UrlAddress string `json:"urladdress"` // url地址
}

type Genre struct {
	ID                  string        `json:"id"`                  // ID号
	N                   int           `json:"n"`                   // 自增序号
	Name                string        `json:"name"`                // 流派名
	AlphabetIndex       string        `json:"alphabetindex"`       // 名字字母表索引
	Description         string        `json:"description"`         // 简介
	Chronology          string        `json:"chronology"`          // 所属年代
	PoetAddresses       []PoetAddress `json:"poetaddresses"`       // 诗人网页地址（最后的路径）
	TimeStamp           string        `json:"timestamp"`           // 创建时间戳
	LastUpdateTimeStamp string        `json:"lastupdatetimestamp"` // 最后更新时间戳
}
