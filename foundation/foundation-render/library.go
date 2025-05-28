package foundationrender

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	html2markdown "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strings"
)

func Render(template string, data map[string]string) string {
	for key, value := range data {
		placeholder := "{{" + key + "}}"
		template = strings.ReplaceAll(template, placeholder, value)
	}
	return template
}

func innerHTML(n *goquery.Selection) string {
	//for c := n.FirstChild; c != nil; c = c.NextSibling {
	//	html.Render(&b, c)
	//}
	return n.Text()
}

func innerText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var result string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result += innerText(c)
	}
	return result
}

func fixAndUploadAllLinks(htmlStr, baseURL string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", err
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, attr := range n.Attr {
				if attr.Key == "src" || attr.Key == "href" {
					original := attr.Val
					if !strings.HasPrefix(original, "http") {
						original = baseURL + original
					}
					uploadedUrl, err := uploadOrPassThrough(original)
					if err == nil {
						n.Attr[i].Val = uploadedUrl
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	var b strings.Builder
	if err := html.Render(&b, doc); err != nil {
		return "", err
	}
	return b.String(), nil
}

func uploadOrPassThrough(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(data)
	hashHex := hex.EncodeToString(hash[:])
	r2Key := fmt.Sprintf("some/path/%s", hashHex)

	// TODO: 上传 data 到你的 R2，然后构建返回 URL
	return fmt.Sprintf("https://r2-oj.boiltask.com/%s", r2Key), nil
}

func HTMLToMarkdown(htmlStr string, baseURL string) (string, error) {
	// Step 1: fix and upload all links
	fixedHtml, err := fixAndUploadAllLinks(htmlStr, baseURL)
	if err != nil {
		return "", err
	}

	// Step 2: convert to markdown
	converter := html2markdown.NewConverter(baseURL, true, nil)
	converter.AddRules(html2markdown.Rule{
		Filter: []string{"pre"},
		Replacement: func(content string, selection *goquery.Selection, opt *html2markdown.Options) *string {
			// 去掉首尾换行
			content = strings.Trim(content, "\n")
			// 去掉内部 div 包裹（若存在）
			if selection.Children().Length() == 1 && goquery.NodeName(selection.Children().First()) == "div" {
				content = strings.Trim(selection.Children().First().Text(), "\n")
			}
			result := fmt.Sprintf("```\n%s\n```", content)
			return &result
		},
	})
	markdown, err := converter.ConvertString(fixedHtml)
	if err != nil {
		return "", err
	}
	return markdown, nil
}
