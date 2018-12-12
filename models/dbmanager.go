package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/mgo.v2"
)

// 数据库信息
const (
	PoemCollection  = "poem"     // 诗歌数据集
	Address         = "address"  // 访问过的url地址
	PoetCollection  = "poet"     // 诗人数据集
	GenreCollection = "genre"    // 诗歌流派
	FailPage        = "failpage" // 抓取失败的地址
)

// MongoConfig 定义数据库配置数据结构
type MongoConfig struct {
	IP        string
	Port      int
	PoolLimit int
	DB        string
}

// UserConfig 定义用户配置数据结构
type UserConfig struct {
	Name string
	Pwd  string
}

// UserUATConfig 定义用户配置数据结构
type UserUATConfig struct {
	Name string
	Pwd  string
}

// Config 定义总配置数据结构
type Config struct {
	Mongo   MongoConfig
	User    UserConfig
	UserUAT UserUATConfig
}

// loadJson 通用读取json文件，并反序列化
func loadJson(filename string, v interface{}) {
	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, v)
	if err != nil {
		return
	}
}

const (
	CaoMangProd = "prodconfig.json"
	CaoMangUAT  = "uatconfig.json"
	SouYun      = "souyun.json"
)

// loadConfig 读取配置文件并反序列化为Config对象
func loadConfig(configName string) Config {
	config := Config{}
	loadJson(configName, &config)
	return config
}

// DBManager 数据库管理器
type DBManager struct {
	Session *mgo.Session
}

// 配置数据
var CONFIG Config

// 数据库URI
var mongoURI string

func init() {
	CONFIG = loadConfig(SouYun)

	mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%v/%s",
		CONFIG.User.Name,
		CONFIG.User.Pwd,
		CONFIG.Mongo.IP,
		CONFIG.Mongo.Port,
		CONFIG.Mongo.DB)

	log.Println("数据库URI：", mongoURI)
}

// NewManager 创建数据库管理器对象
func NewDBManager() (*DBManager, error) {
	Session, err := mgo.Dial(mongoURI)
	if err != nil {
		return nil, err
	}
	return &DBManager{Session}, nil
}

// SetDB 根据数据库名字，创建数据库连接
func (m *DBManager) SetDB(name string) *mgo.Database {
	return m.Session.DB(name)
}

// Coll 根据数据库表名，返回表对象
func (m *DBManager) Coll(name string) *mgo.Collection {
	return m.Session.DB(CONFIG.Mongo.DB).C(name)
}

// Close 关闭数据库
func (m *DBManager) Close() {
	m.Session.Close()
}
