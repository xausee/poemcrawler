package models

import (
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Poet 诗人数据结构
type Poet struct {
	ID                  string   // ID号
	Name                string   // 名字
	AlphabetIndex       string   // 名字字母表索引
	Chronology          string   // 诗人所属年代
	Dynasty             string   // 诗人朝代，用于古代诗人，近代诗人使用上面Chronology（年代）字段
	Genres              []string // 诗人所属流派，可以是多个
	Intro               string   // 简介
	Avatar              string   // 头像：ID号.png
	Source              string   // 来源
	TimeStamp           string   // 创建时间戳
	LastUpdateTimeStamp string   // 最后更新时间戳
}

// PoetModel 诗人数据库model
type PoetModel struct {
	db   *DBManager      // 数据库对象
	coll *mgo.Collection // 诗人对应的数据集
}

// NewPoetModel 创建诗人Model对象
// 使用完后务必调用"Dispose()"以关闭数据库会话
func NewPoetModel() *PoetModel {
	db, e := NewDBManager()
	if e != nil {
		log.Println(e)
		return nil
	}

	//defer db.Close()

	c := db.Session.DB(CONFIG.Mongo.DB).C(PoetCollection)
	return &PoetModel{db: db, coll: c}
}

// Dispose 释放对象资源
func (m PoetModel) Dispose() {
	m.db.Close()
}

// All 获取所有诗人信息
func (m PoetModel) All() ([]Poet, error) {
	var r []Poet

	e := m.coll.Find(bson.M{}).All(&r)

	return r, e
}

// AllWithDefaultAvatar 获取所有还是缺省头像的诗人信息
func (m PoetModel) AllWithDefaultAvatar() ([]Poet, error) {
	var a []Poet

	e := m.coll.Find(bson.M{"avatar": ""}).All(&a)

	return a, e
}

// FindByID 根据ID查找诗人信息
func (m PoetModel) FindByID(id string) (Poet, error) {
	var r Poet

	e := m.coll.Find(bson.M{"id": id}).One(&r)

	return r, e
}

// UpdateIntro 更新描述
func (m PoetModel) UpdateIntro(id, intro string) error {
	var o Poet

	e := m.coll.Find(bson.M{"id": id}).One(&o)
	if e != nil {
		return e
	}

	n := o
	n.Intro = intro

	e = m.coll.Update(o, n)
	if e != nil {
		return e
	}

	return nil
}

// UpdateAvatar 更新头像
func (m PoetModel) UpdateAvatar(id, avatar string) error {
	var o Poet

	e := m.coll.Find(bson.M{"id": id}).One(&o)
	if e != nil {
		return e
	}

	n := o
	n.Avatar = avatar

	e = m.coll.Update(o, n)
	if e != nil {
		return e
	}

	return nil
}
