package foundationrender

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	html2markdown "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"io"
	metahttp "meta/meta-http"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

func Render(template string, data map[string]string) string {
	for key, value := range data {
		placeholder := "{{" + key + "}}"
		template = strings.ReplaceAll(template, placeholder, value)
	}
	return template
}

type Replacement struct {
	Start       int
	End         int
	Replacement string
	ForceUpload bool
}

func isAbsoluteUrl(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

// 模拟上传到 R2，实际请替换成自己的逻辑
func uploadToR2(problemKey string, content []byte) (string, error) {
	hash := md5.Sum(content)
	hashHex := hex.EncodeToString(hash[:])
	r2Key := metahttp.UrlJoin(problemKey, hashHex)

	// 模拟上传逻辑
	// 实际替换成 AWS S3 SDK / Cloudflare R2 SDK 上传
	fmt.Println("Uploading to R2:", r2Key)

	// 返回访问 URL
	return metahttp.UrlJoin("https://r2-oj.boiltask.com", r2Key), nil
}

func fixAndUploadAllLinks(problemKey string, html string, baseUrl string) (string, error) {
	var replacements []Replacement

	collectLinks := func(pattern *regexp.Regexp, forceUpload bool) {
		matches := pattern.FindAllStringSubmatchIndex(html, -1)
		for _, match := range matches {
			start := match[0]
			end := match[1]
			replacements = append(replacements, Replacement{
				Start:       start,
				End:         end,
				Replacement: "",
				ForceUpload: forceUpload,
			})
		}
	}

	// 收集资源链接
	collectLinks(regexp.MustCompile(`src=["']([^"']+)["']`), true)
	collectLinks(regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)`), true)
	collectLinks(regexp.MustCompile(`href=["']([^"']+)["']`), false)

	// 并发处理替换
	for i := range replacements {
		entry := &replacements[i]
		raw := html[entry.Start:entry.End]

		pathMatch := regexp.MustCompile(`["']([^"']+)["']`).FindStringSubmatch(raw)
		if pathMatch == nil {
			pathMatch = regexp.MustCompile(`\(([^)]+)\)`).FindStringSubmatch(raw)
		}
		if pathMatch == nil {
			continue
		}
		originalPath := pathMatch[1]
		var absoluteUrl string
		if isAbsoluteUrl(originalPath) {
			absoluteUrl = originalPath
		} else {
			absoluteUrl = metahttp.UrlJoin(baseUrl, originalPath)
		}

		var newUrl string
		if !entry.ForceUpload && !regexp.MustCompile(`\.(png|jpe?g|gif|webp)$`).MatchString(absoluteUrl) {
			newUrl = absoluteUrl
		} else {
			resp, err := http.Get(absoluteUrl)
			if err != nil {
				return "", fmt.Errorf("failed to download %s: %w", absoluteUrl, err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("failed to read body: %w", err)
			}

			newUrl, err = uploadToR2(problemKey, body)
			if err != nil {
				return "", fmt.Errorf("upload failed: %w", err)
			}
		}

		if strings.HasPrefix(raw, "src=") || strings.HasPrefix(raw, "href=") {
			attr := strings.Split(raw, "=")[0]
			entry.Replacement = fmt.Sprintf(`%s="%s"`, attr, newUrl)
		} else {
			altMatch := regexp.MustCompile(`!\[([^\]]*)\]`).FindStringSubmatch(raw)
			alt := ""
			if len(altMatch) >= 2 {
				alt = altMatch[1]
			}
			entry.Replacement = fmt.Sprintf(`![%s](%s)`, alt, newUrl)
		}
	}

	// 排序并拼接新的 HTML
	sort.Slice(replacements, func(i, j int) bool {
		return replacements[i].Start < replacements[j].Start
	})

	var buf bytes.Buffer
	lastIndex := 0
	for _, r := range replacements {
		buf.WriteString(html[lastIndex:r.Start])
		buf.WriteString(r.Replacement)
		lastIndex = r.End
	}
	buf.WriteString(html[lastIndex:])

	return buf.String(), nil
}

func HTMLToMarkdown(problemId string, htmlStr string, baseURL string) (string, error) {
	// Step 1: fix and upload all links
	fixedHtml, err := fixAndUploadAllLinks(problemId, htmlStr, baseURL)
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
