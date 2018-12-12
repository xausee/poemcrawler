package main

import (
	"PoemCrawler/dpc"
	"PoemCrawler/ht"
	"PoemCrawler/saver"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/PuerkitoBio/gocrawl"
	"github.com/PuerkitoBio/goquery"
)

type Ext struct {
	*gocrawl.DefaultExtender
}

func (e *Ext) Visit(ctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) (interface{}, bool) {
	log.Println("访问 " + ctx.URL().String())
	//必须要先声明defer，否则不能捕获到panic异常
	defer func() {
		if err := recover(); err != nil {
			log.Println("处理 " + ctx.URL().String() + " 时产生异常，继续...")
			debug.PrintStack()
			//e.Visit(ctx, res, doc)
			db.SaveFailPage(ctx.URL().String())
		}
	}()

	d := dpc.NewDispatcher(ctx, res, doc)
	if ctx.URL().Host == "sou-yun.com" {
		d.DispatchToSouYun()
	}

	if ctx.URL().Host == "www.shiku.org" {
		d.DispatchToShiKu()
	}

	return nil, true
}

func (e *Ext) Filter(ctx *gocrawl.URLContext, isVisited bool) bool {
	if isVisited {
		return false
	}

	// 中华诗库古典诗库 页面太不规范，不再爬取
	if strings.Contains(ctx.URL().String(), "http://www.shiku.org/shiku/gs") &&
		strings.Contains(ctx.URL().Path, ".htm") {
		return false
	}

	// 中华诗库现代诗库
	if strings.Contains(ctx.URL().String(), "http://www.shiku.org/shiku/xs") &&
		strings.Contains(ctx.URL().Path, ".htm") {
		return true
	}

	// 中华诗库国际诗库
	if strings.Contains(ctx.URL().String(), "http://www.shiku.org/shiku/ws") &&
		strings.Contains(ctx.URL().Path, ".htm") {
		return true
	}

	//if ctx.URL().Host == "www.shiku.org" && (strings.Contains(ctx.URL().Path, ".htm") ||
	//	strings.Contains(ctx.URL().Path, ".html")) {
	//	return true
	//}

	// 以下是搜韵过滤条件
	if strings.Contains(ctx.URL().String(), "&dm=") || strings.Contains(ctx.URL().String(), "&page=0") {
		return false
	}

	if strings.Contains(ctx.URL().String(), "&type=") && !(strings.Contains(ctx.URL().String(), "	&type=All") ||
		strings.Contains(ctx.URL().String(), "&type=All&page=")) {
		return false
	}

	if ctx.URL().Host == "sou-yun.com" && strings.Contains(ctx.URL().String(), "PoemIndex.aspx?dynasty=") {
		return true
	}
	// 搜韵过滤条件结束

	return false
}

func main() {
	ext := &Ext{&gocrawl.DefaultExtender{}}
	// Set custom options
	opts := gocrawl.NewOptions(ext)
	opts.CrawlDelay = 1 * time.Second
	opts.LogFlags = gocrawl.LogError
	opts.SameHostOnly = false
	opts.MaxVisits = 1000000
	//opts.MaxVisits = 1

	c := gocrawl.NewCrawlerWithOptions(opts)
	// 中华诗库总目录
	// c.Run("http://www.shiku.org/shiku/index.htm")
	// 现代诗库
	c.Run("http://www.shiku.org/shiku/xs/index.htm")
	// 古典诗库
	// c.Run("http://www.shiku.org/shiku/gs/index.htm")
	// 国际诗库
	// c.Run("http://www.shiku.org/shiku/ws/index.htm")
	// 搜韵网
	// c.Run("https://sou-yun.com/PoemIndex.aspx?dynasty=XianQin")

    // 爬取完毕后统一处理爬取失败的页面（再爬取一次）
	ht.ParseAllFailPage()
}
