package models

import (
	"log"
	"math/rand"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Comment 评论
type Comment struct {
	ID          string // 评论ID号
	CommenterID string // 评论人ID
	Body        string // 评论内容
	TimeStamp   string // 时间戳
}

// PoemType 诗歌类型
type PoemType int

// iota 初始化后会自动递增
const (
	XianDai PoemType = iota // value --> 0
	GuDian                  // value --> 1
	YiShi                   // value --> 2
	WaiWen                  // value --> 3
	WeiZhi                  // value --> 4
)

// String String方法
func (t PoemType) String() string {
	switch t {
	case XianDai:
		return "现代"
	case GuDian:
		return "古典"
	case YiShi:
		return "译诗"
	case WaiWen:
		return "外文"
	case WeiZhi:
		return "未知"
	default:
		return "未知"
	}
}

// Poem 结构
type Poem struct {
	N                   int      // 自增序号
	ID                  string   // ID号
	Author              string   // 作者名字
	AuthorID            string   // 作者ID号
	Dynasty             string   // 朝代
	Cover               string   // 封面，没有实际值存到数据库，用于数据传输到前端时存储诗人头像地址
	Type                string   // 诗歌类型：现代、古典、译诗、外文
	Tags                []string // 标签
	Title               string   // 标题
	SubTitle            string   // 副标题
	Volume              string   // 卷
	Content             string   // 内容
	Commentary          string   // 注释、注解
	Source              string   // 来源
	View                int      // 阅读次数
	Praise              int      // 点赞次数
	CommentID           []string // 评论ID集合
	TimeStamp           string   // 创建时间戳
	LastUpdateTimeStamp string   // 最后更新时间戳
}

// PoemModel 诗歌数据库model
type PoemModel struct {
	db   *DBManager      // 数据库对象
	coll *mgo.Collection // 诗人对应的数据集
}

// NewPoemModel 创建诗人Model对象
// 使用完后务必调用"Dispose()"以关闭数据库会话
func NewPoemModel() *PoemModel {
	db, e := NewDBManager()
	if e != nil {
		log.Println(e)
		return nil
	}

	//defer db.Close()

	c := db.Session.DB(config.Mongo.DB).C(PoemCollection)
	return &PoemModel{db: db, coll: c}
}

// Dispose 释放对象资源
func (m PoemModel) Dispose() {
	m.db.Close()
}

// FindLast 查找最新的n个记录
func (m PoemModel) FindLast(n int) ([]Poem, error) {
	var r []Poem

	e := m.coll.Find(bson.M{}).Sort("-createtimestamp").Limit(n).All(&r)

	return r, e
}

// Search 查找最新的n个记录
func (m PoemModel) Search(k string) ([]Poem, error) {
	var r []Poem

	q := bson.M{
		"$or": []bson.M{
			bson.M{"author": bson.M{"$regex": k}},
			bson.M{"title": bson.M{"$regex": k}}},
	}

	e := m.coll.Find(q).Sort("-createtimestamp").Limit(20).All(&r)

	return r, e
}

// FindRandom 随机获取n个诗歌记录
func (m PoemModel) FindRandom(n int) ([]Poem, error) {
	var r []Poem

	if n == 0 {
		return r, nil
	}

	rand.Seed(time.Now().UnixNano())

	c, e := m.coll.Find(bson.M{}).Count()
	if e != nil || c == 0 {
		return r, e
	}

	for i := 0; i < n; i++ {
		var p Poem
		e = m.coll.Find(bson.M{}).Skip(rand.Intn(c)).One(&p)
		if e != nil {
			log.Println(e)
		} else {
			r = append(r, p)
		}
	}

	return r, nil
}

// FindRandomWithType1 随机获取n个t类型的诗歌记录
func (m PoemModel) FindRandomWithType1(n int, t string) ([]Poem, error) {
	var r []Poem

	if n == 0 {
		return r, nil
	}

	rand.Seed(time.Now().UnixNano())

	c, e := m.coll.Find(bson.M{"type": t}).Count()
	if e != nil || c == 0 {
		return r, e
	}

	for i := 0; i < n; i++ {
		var p Poem
		e = m.coll.Find(bson.M{"type": t}).Skip(rand.Intn(c)).One(&p) // Skip 有性能问题
		if e != nil {
			log.Println(e)
		} else {
			r = append(r, p)
		}
	}

	return r, nil
}

// FindRandomWithType 随机获取n个t类型的诗歌记录
func (m PoemModel) FindRandomWithType(n int, t string) ([]Poem, error) {
	var r []Poem

	if n == 0 {
		return r, nil
	}

	rand.Seed(time.Now().UnixNano())

	c, e := m.coll.Find(bson.M{"type": t}).Count()
	if e != nil || c == 0 {
		return r, e
	}

	a := make([]int, 0, 0)
	for i := 0; i < n; i++ {
		a = append(a, rand.Intn(c))
	}
	e = m.coll.Find(bson.M{"type": t, "n": bson.M{"$in": a}}).All(&r)
	if e != nil {
		log.Println(e)
	}

	return r, nil
}

// FindByID 根据id查找诗歌记录
func (m PoemModel) FindByID(id string) (Poem, error) {
	var r Poem

	e := m.coll.Find(bson.M{"id": id}).One(&r)

	return r, e
}
