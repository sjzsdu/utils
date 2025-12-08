package sources

import (
	"encoding/json"
	"net/url"
	"regexp"
	"time"

	"github.com/sjzsdu/utils/crawler/pkg/models"
)

// KuaishouSource 快手数据源
// 该数据源从快手首页HTML中提取内嵌的热榜数据
type KuaishouSource struct {
	BaseSource
}

// NewKuaishouSource 创建快手数据源实例
func NewKuaishouSource() *KuaishouSource {
	return &KuaishouSource{
		BaseSource: BaseSource{
			Name:       "kuaishou",
			URL:        "https://www.kuaishou.com/",
			Interval:   300, // 5分钟爬取一次
			Categories: []string{"娱乐", "综合"},
		},
	}
}

// Parse 解析快手热榜内容
func (s *KuaishouSource) Parse(content []byte) ([]models.Item, error) {
	// 从HTML中提取window.__APOLLO_STATE__的数据
	re := regexp.MustCompile(`window\.__APOLLO_STATE__\s*=\s*(\{.+?\});`)
	matches := re.FindSubmatch(content)
	if len(matches) < 2 {
		return nil, nil
	}

	// 解析JSON数据
	var apolloState map[string]interface{}
	if err := json.Unmarshal(matches[1], &apolloState); err != nil {
		return nil, err
	}

	// 获取defaultClient
	defaultClient, ok := apolloState["defaultClient"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	// 获取ROOT_QUERY
	rootQuery, ok := defaultClient["ROOT_QUERY"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	// 获取热榜数据ID
	hotRankKey := "visionHotRank({\"page\":\"home\"})"
	hotRankInfo, ok := rootQuery[hotRankKey].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	hotRankID, ok := hotRankInfo["id"].(string)
	if !ok {
		return nil, nil
	}

	// 获取热榜列表数据
	hotRankData, ok := defaultClient[hotRankID].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	// 获取items数组
	itemsInterface, ok := hotRankData["items"].([]interface{})
	if !ok {
		return nil, nil
	}

	var items []models.Item
	// 处理每个热榜项
	for _, itemInterface := range itemsInterface {
		item, ok := itemInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// 获取item的id
		itemID, ok := item["id"].(string)
		if !ok {
			continue
		}

		// 获取具体的热榜项数据
		hotItem, ok := defaultClient[itemID].(map[string]interface{})
		if !ok {
			continue
		}

		// 过滤置顶内容
		tagType, ok := hotItem["tagType"].(string)
		if ok && tagType == "置顶" {
			continue
		}

		// 获取名称
		name, ok := hotItem["name"].(string)
		if !ok {
			continue
		}

		// 从id中提取热搜词
		hotSearchWord := itemID
		if len(itemID) > 18 && itemID[:18] == "VisionHotRankItem:" {
			hotSearchWord = itemID[18:]
		}

		// 构建URL
		searchURL := "https://www.kuaishou.com/search/video?searchKey=" + url.QueryEscape(name)

		// 构建extra字段
		extra := make(map[string]interface{})
		if iconURL, ok := hotItem["iconUrl"].(string); ok && iconURL != "" {
			extra["icon"] = iconURL
		}

		items = append(items, models.Item{
			ID:          hotSearchWord,
			Title:       name,
			URL:         searchURL,
			Source:      s.Name,
			CreatedAt:   time.Now(),
			PublishedAt: time.Now(),
		})
	}

	return items, nil
}

func init() {
	RegisterSource(NewKuaishouSource())
}
