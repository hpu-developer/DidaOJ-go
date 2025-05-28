package service

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"meta/singleton"
	"net/http"
	"os"
	"strings"
)

// 定义结构体
type ProblemDetailResponse struct {
	Status int `json:"status"`
	Data   struct {
		Problem struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Input       string `json:"input"`
			Output      string `json:"output"`
			Examples    string `json:"examples"`
			Source      string `json:"source"`
			Hint        string `json:"hint"`
		} `json:"problem"`
	} `json:"data"`
}

// 工具函数：清理 HTML 字符实体 + 替换 <p> 等为换行
func cleanHTML(s string) string {
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "<p>", "")
	s = strings.ReplaceAll(s, "</p>", "\n")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimSpace(s)
}

// 将题目信息转为 Markdown 格式
func toMarkdown(p ProblemDetailResponse) string {
	prob := p.Data.Problem

	// 处理输入输出示例（可能是 HTML 包装的 input/output）
	examples := html.UnescapeString(prob.Examples)
	input, output := "", ""
	if strings.Contains(examples, "<input>") && strings.Contains(examples, "</input>") {
		input = extractBetween(examples, "<input>", "</input>")
	}
	if strings.Contains(examples, "<output>") && strings.Contains(examples, "</output>") {
		output = extractBetween(examples, "<output>", "</output>")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", prob.Title))
	sb.WriteString("## 题目描述\n\n")
	sb.WriteString(cleanHTML(prob.Description) + "\n\n")

	sb.WriteString("## 输入格式\n\n")
	sb.WriteString(cleanHTML(prob.Input) + "\n\n")

	sb.WriteString("## 输出格式\n\n")
	sb.WriteString(cleanHTML(prob.Output) + "\n\n")

	if input != "" || output != "" {
		sb.WriteString("## 样例输入\n\n```\n" + input + "\n```\n\n")
		sb.WriteString("## 样例输出\n\n```\n" + output + "\n```\n\n")
	}

	if prob.Hint != "" {
		sb.WriteString("## 提示\n\n" + cleanHTML(prob.Hint) + "\n\n")
	}

	sb.WriteString("## 数据来源\n\n" + prob.Source + "\n")

	return sb.String()
}

// 提取 HTML 标签内的内容
func extractBetween(s, start, end string) string {
	i := strings.Index(s, start)
	j := strings.Index(s, end)
	if i >= 0 && j >= 0 && j > i+len(start) {
		return s[i+len(start) : j]
	}
	return ""
}

type CrawlNyojService struct {
}

var singletonCrawlNyojService = singleton.Singleton[CrawlNyojService]{}

func GetCrawlNyojService() *CrawlNyojService {
	return singletonCrawlNyojService.GetInstance(
		func() *CrawlNyojService {
			return &CrawlNyojService{}
		},
	)
}

func (s *CrawlNyojService) PostCrawlProblem(ctx context.Context, id string) (*string, error) {
	s.RunNyoj(id)
	return nil, nil
}

func (s *CrawlNyojService) RunNyoj(problemId string) {
	url := fmt.Sprintf("https://xcpc.nyist.edu.cn/api/get-problem-detail?problemId=%s", problemId)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("状态码错误: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var detail ProblemDetailResponse
	if err := json.Unmarshal(body, &detail); err != nil {
		log.Fatal(err)
	}

	md := toMarkdown(detail)
	filename := fmt.Sprintf("problem_%s.md", problemId)
	if err := os.WriteFile(filename, []byte(md), 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("已保存 Markdown 文件: %s\n", filename)
}
