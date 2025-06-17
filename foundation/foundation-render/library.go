package foundationrender

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/strikethrough"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
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
			replacements = append(
				replacements, Replacement{
					Start:       start,
					End:         end,
					Replacement: "",
					ForceUpload: forceUpload,
				},
			)
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
	sort.Slice(
		replacements, func(i, j int) bool {
			return replacements[i].Start < replacements[j].Start
		},
	)

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

func extractMathBlocks(html string) (string, map[string]string) {

	re := regexp.MustCompile(`\$(.+?)\$`)
	matches := re.FindAllStringSubmatch(html, -1)

	placeholderMap := make(map[string]string)
	for i, match := range matches {
		key := fmt.Sprintf(`{{MARKDOWN-MATH::%d}}"`, i)
		placeholderMap[key] = match[0]
		html = strings.Replace(html, match[0], key, 1)
	}

	return html, placeholderMap
}

func restoreMathBlocks(markdown string, placeholderMap map[string]string) string {
	for key, val := range placeholderMap {
		markdown = strings.Replace(markdown, key, val, 1)
	}
	return markdown
}

func HTMLToMarkdown(problemId string, htmlStr string, baseURL string) (string, error) {

	preprocessedHTML, mathMap := extractMathBlocks(htmlStr)

	// Step 1: fix and upload all links
	fixedHtml, err := fixAndUploadAllLinks(problemId, preprocessedHTML, baseURL)
	if err != nil {
		return "", err
	}

	// Step 2: convert to markdown
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			strikethrough.NewStrikethroughPlugin(),
			table.NewTablePlugin(),
		),
	)
	markdown, err := conv.ConvertString(fixedHtml, converter.WithDomain(baseURL))
	if err != nil {
		return "", err
	}

	// 3. 恢复 $$...$$
	finalMarkdown := restoreMathBlocks(markdown, mathMap)

	return finalMarkdown, nil
}
