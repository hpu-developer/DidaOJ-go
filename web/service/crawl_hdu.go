package service

import (
	"bytes"
	"context"
	"fmt"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	foundationrender "foundation/foundation-render"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"log/slog"
	metaerror "meta/meta-error"
	metapanic "meta/meta-panic"
	metatime "meta/meta-time"
	"meta/singleton"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"web/config"

	"github.com/PuerkitoBio/goquery"
)

type CrawlHduService struct {
}

var singletonCrawlHduService = singleton.Singleton[CrawlHduService]{}

func GetCrawlHduService() *CrawlHduService {
	return singletonCrawlHduService.GetInstance(
		func() *CrawlHduService {
			return &CrawlHduService{}
		},
	)
}

func (s *CrawlHduService) PostCrawlProblem(ctx context.Context, id string) (*string, error) {

	nowTime := metatime.GetTimeNow()
	newProblemId := fmt.Sprintf("HDU-%s", id)
	baseURL := "https://acm.hdu.edu.cn/"
	url := fmt.Sprintf("%sshowproblem.php?pid=%s", baseURL, id)

	// 获取原始 HTML
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch problem page")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err))
		}
	}(resp.Body)

	// 解码 GBK 到 UTF-8
	gbkData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	reader := transform.NewReader(bytes.NewReader(gbkData), simplifiedchinese.GBK.NewDecoder())
	utf8Html, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// 解析 HTML
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(utf8Html))
	if err != nil {
		return nil, err
	}

	// 标题
	title := strings.TrimSpace(doc.Find("h1").First().Text())

	// panel_title / panel_content 处理
	contentMap := make(map[string]string)
	var currentSection string
	var finalErr error
	doc.Find(".panel_title, .panel_content").Each(func(i int, s *goquery.Selection) {
		if finalErr != nil {
			return // 如果已经有错误了，就不再处理
		}
		if s.HasClass("panel_title") {
			currentSection = strings.TrimSpace(s.Text())
			// 如果需要 markdown 转换，这里调用自定义函数
		} else if s.HasClass("panel_content") {
			htmlContent, _ := s.Html()
			htmlContent, err = foundationrender.HTMLToMarkdown(newProblemId, htmlContent, baseURL)
			if err != nil {
				finalErr = metaerror.Join(finalErr, err)
				return
			}
			contentMap[currentSection] = htmlContent
		}
	})
	if finalErr != nil {
		return nil, finalErr
	}

	// 时间/内存限制
	tbodyText := doc.Find("tbody").Text()
	re := regexp.MustCompile(`Time Limit: \d+/(\d+) MS \(Java/Others\)[\s\S]*Memory Limit: \d+/(\d+) K \(Java/Others\)`)
	matches := re.FindStringSubmatch(tbodyText)
	timeLimit, memoryLimit := -1, -1
	if len(matches) >= 3 {
		timeLimitStr := strings.TrimSpace(matches[1])
		memoryLimitStr := strings.TrimSpace(matches[2])
		timeLimit, err = strconv.Atoi(timeLimitStr)
		if err != nil {
			return nil, metaerror.Wrap(err, "parse time limit failed")
		}
		memoryLimit, err = strconv.Atoi(memoryLimitStr)
		if err != nil {
			return nil, metaerror.Wrap(err, "parse memory limit failed")
		}
	}

	author := contentMap["Author"]
	source := contentMap["Source"]

	hint, ok := contentMap["Hint"]
	if ok {
		hint = fmt.Sprintf("\n\n## Hint\n\n%s", hint)
	}

	// 渲染模板（这里简化处理）
	template := config.GetOjTemplateContent("hdu")
	description := foundationrender.Render(template, map[string]string{
		"description":  contentMap["Problem Description"],
		"input":        contentMap["Input"],
		"output":       contentMap["Output"],
		"sampleInput":  contentMap["Sample Input"],
		"sampleOutput": contentMap["Sample Output"],
		"hint":         hint,
	})

	slog.Info("description", description)

	originUrl := fmt.Sprintf("https://acm.hdu.edu.cn/showproblem.php?pid=%s", id)

	problem := foundationmodel.NewProblemBuilder().
		Id(newProblemId).
		Sort(len(newProblemId)).
		Title(title).
		Description(description).
		TimeLimit(timeLimit).
		MemoryLimit(memoryLimit).
		CreatorNickname(author).
		Source(source).
		OriginOj("HDU").
		OriginId(id).
		OriginUrl(originUrl).
		InsertTime(nowTime).
		UpdateTime(nowTime).
		Build()

	err = foundationdao.GetProblemDao().UpdateProblemCrawl(ctx, newProblemId, problem)
	if err != nil {
		return nil, err
	}

	return &newProblemId, nil
}

// optionalSection 帮助渲染 hint/author/source
func optionalSection(name string, m map[string]string) string {
	if v, ok := m[name]; ok && v != "" {
		return fmt.Sprintf("\n## %s\n\n%s", name, v)
	}
	return ""
}
