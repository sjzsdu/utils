package parser

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// HTMLParser 是HTML解析器
type HTMLParser struct{}

// NewHTMLParser 创建一个新的HTML解析器实例
func NewHTMLParser() *HTMLParser {
	return &HTMLParser{}
}

// ParseElement 解析HTML内容，提取指定元素
func (p *HTMLParser) ParseElement(content []byte, selector string) (*goquery.Selection, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		return nil, err
	}

	return doc.Find(selector), nil
}

// ExtractText 提取指定选择器的文本内容
func (p *HTMLParser) ExtractText(content []byte, selector string) (string, error) {
	selection, err := p.ParseElement(content, selector)
	if err != nil {
		return "", err
	}

	return selection.Text(), nil
}

// ExtractAttr 提取指定选择器的属性值
func (p *HTMLParser) ExtractAttr(content []byte, selector, attr string) (string, error) {
	selection, err := p.ParseElement(content, selector)
	if err != nil {
		return "", err
	}

	return selection.AttrOr(attr, ""), nil
}

// ExtractLinks 提取指定选择器下的所有链接
func (p *HTMLParser) ExtractLinks(content []byte, selector string) ([]string, error) {
	selection, err := p.ParseElement(content, selector)
	if err != nil {
		return nil, err
	}

	var links []string
	selection.Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			links = append(links, href)
		}
	})

	return links, nil
}

// ExtractImages 提取指定选择器下的所有图片链接
func (p *HTMLParser) ExtractImages(content []byte, selector string) ([]string, error) {
	selection, err := p.ParseElement(content, selector)
	if err != nil {
		return nil, err
	}

	var images []string
	selection.Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			images = append(images, src)
		}
	})

	return images, nil
}

// FindAll 查找所有匹配选择器的元素
func (p *HTMLParser) FindAll(content []byte, selector string) ([]*goquery.Selection, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		return nil, err
	}

	var selections []*goquery.Selection
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		selections = append(selections, s)
	})

	return selections, nil
}
