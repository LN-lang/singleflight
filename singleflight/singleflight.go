package singleflight

import (
	"context"
	"fmt"

	"github.com/go-rod/rod"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

var (
	rdb     = redis.NewClient(&redis.Options{Addr: " 101.126.43.173:6379 "})
	sfGroup singleflight.Group
	ctx     = context.Background()
)

// 核心：热点数据获取（缓存+请求合并）
func GetHotData(key string) string {
	// 1. 查缓存
	if v, _ := rdb.Get(ctx, key).Result(); v != "" {
		return v
	}
	// 2. 合并请求查数据源+写缓存
	val, _, _ := sfGroup.Do(key, func() (interface{}, error) {
		// ========== 替换这里：你的爬虫逻辑 ==========
		browser := rod.New().MustConnect()
		defer browser.MustClose()
		page := browser.MustPage(" https://air.1688.com/kapp/channel-fe/cps-4c-pc/sytm?type=1&offerIds=660390230106,574965204819,949033739317 ")
		page.MustWaitLoad()

		// 提取商品名称+价格（核心爬取逻辑）
		name, _ := page.MustElementX("//*[@id=\"ice-container\"]/div/div/div[3]/a[1]/div/div[1]").Text()
		price, _ := page.MustElementX("//*[@id=\"ice-container\"]/div/div/div[3]/a[1]/div/div[3]").Text()
		data := fmt.Sprintf("商品：%s | 价格：%s", name, price)
		// ==========================================

		rdb.Set(ctx, key, data, 5*60) // 缓存5分钟
		return data, nil
	})
	return val.(string)
}
