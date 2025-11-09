package foundationremote

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	foundationconfig "foundation/foundation-config"
	foundationdao "foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationrender "foundation/foundation-render"
	"io"
	"log/slog"
	metaerror "meta/meta-error"
	metapanic "meta/meta-panic"
	metatime "meta/meta-time"
	"meta/singleton"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/PuerkitoBio/goquery"
)

type RemotePojAgent struct {
	goJudgeClient *http.Client

	cookie string

	mutex sync.Mutex
}

var singletonRemotePojAgent = singleton.Singleton[RemotePojAgent]{}

func GetRemotePojAgent() *RemotePojAgent {
	return singletonRemotePojAgent.GetInstance(
		func() *RemotePojAgent {
			s := &RemotePojAgent{}

			s.goJudgeClient = &http.Client{
				Transport: &http.Transport{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 100,
					MaxConnsPerHost:     100,
					IdleConnTimeout:     90 * time.Second,
				},
				Timeout: 60 * time.Second, // 请求整体超时
			}

			return s
		},
	)
}

func (s *RemotePojAgent) getLanguageCode(language foundationjudge.JudgeLanguage) string {
	switch language {
	case foundationjudge.JudgeLanguageC:
		return "1"
	case foundationjudge.JudgeLanguageCpp:
		return "0"
	case foundationjudge.JudgeLanguagePascal:
		return "4"
	case foundationjudge.JudgeLanguageJava:
		return "5"
	default:
		return ""
	}
}

func (s *RemotePojAgent) GetJudgeStatus(status string) foundationjudge.JudgeStatus {
	switch status {
	case "Queuing":
		return foundationjudge.JudgeStatusQueuing
	case "Compiling":
		return foundationjudge.JudgeStatusCompiling
	case "Running":
		return foundationjudge.JudgeStatusRunning
	case "Accepted":
		return foundationjudge.JudgeStatusAC
	case "Presentation Error":
		return foundationjudge.JudgeStatusPE
	case "Wrong Answer":
		return foundationjudge.JudgeStatusWA
	case "Runtime Error":
		return foundationjudge.JudgeStatusRE
	case "Time Limit Exceeded":
		return foundationjudge.JudgeStatusTLE
	case "Memory Limit Exceeded":
		return foundationjudge.JudgeStatusMLE
	case "Output Limit Exceeded":
		return foundationjudge.JudgeStatusOLE
	case "Compilation Error":
		return foundationjudge.JudgeStatusCE
	case "System Error":
		return foundationjudge.JudgeStatusRE
	default:
		return foundationjudge.JudgeStatusJudgeFail
	}
}

func (s *RemotePojAgent) IsSupportJudge(problemId string, language foundationjudge.JudgeLanguage) bool {
	return s.getLanguageCode(language) != ""
}

func (s *RemotePojAgent) login(ctx context.Context) error {

	slog.Info("POJ remote judge login start")

	loginUrl := "https://vjudge.net/user/login"
	method := "POST"
	username := foundationconfig.GetConfig().Remote.Hdu.Username
	password := foundationconfig.GetConfig().Remote.Hdu.Password

	payload := strings.NewReader(fmt.Sprintf("username=%s&userpass=%s", username, password))
	req, err := http.NewRequestWithContext(ctx, method, loginUrl, payload)
	if err != nil {
		return metaerror.Wrap(err, "failed to create login request")
	}
	req.Header.Add("priority", "u=1, i")
	req.Header.Add("x-requested-with", "XMLHttpRequest")
	req.Header.Add("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "vjudge.net")
	req.Header.Add("Connection", "keep-alive")

	res, err := s.goJudgeClient.Do(req)
	if err != nil {
		return metaerror.Wrap(err, "login request failed")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err))
		}
	}(res.Body)
	cookie := res.Header.Get("Set-Cookie")
	if cookie == "" {
		return metaerror.New("login failed, no Set-Cookie header")
	}
	cookieParts := strings.SplitN(cookie, ";", 2)
	if len(cookieParts) == 0 {
		return metaerror.New("login failed, invalid Set-Cookie header")
	}
	cookie = strings.TrimSpace(cookieParts[0])
	s.cookie = fmt.Sprintf("JSESSIONlD:%s|%s; Path=/; Domain=vjudge.net;", username, cookie)
	return nil
}

func (s *RemotePojAgent) crawlProblem(ctx context.Context, id string, retryCount int) (*string, error) {

	nowTime := metatime.GetTimeNow()
	newProblemId := fmt.Sprintf("POJ-%s", id)
	baseURL := "https://vjudge.net/"

	vjudgeUrl := fmt.Sprintf("https://vjudge.net/problem/%s", newProblemId)
	method := "GET"
	req, err := http.NewRequestWithContext(ctx, method, vjudgeUrl, nil)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to create PostCrawlProblem request")
	}
	req.Header.Add("priority", "u=0, i")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "vjudge.net")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Username", foundationconfig.GetConfig().Remote.VJudge.Username)
	req.Header.Add("Cookie", s.cookie)

	resp, err := s.goJudgeClient.Do(req)
	if err != nil {
		return nil, metaerror.Wrap(err, "PostCrawlProblem request failed")
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

	descLi := doc.Find(`ul#prob-descs li`).FilterFunction(
		func(i int, s *goquery.Selection) bool {
			return s.Find("b").First().Text() == "System Crawler"
		},
	).First()

	descUrl, _ := descLi.Find(`.operation a[target="_blank"]`).Attr("href")
	if descUrl == "" {
		if retryCount < 1 {
			// 重新登录一次试试
			err := s.login(ctx)
			if err != nil {
				return nil, err
			}
			return s.crawlProblem(ctx, id, retryCount+1)
		}
		return nil, metaerror.New("failed to get problem description URL")
	}
	fullDescUrl := "https://vjudge.net" + descUrl

	descResp, err := http.Get(fullDescUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err))
		}
	}(descResp.Body)

	descGbkData, err := io.ReadAll(descResp.Body)
	if err != nil {
		return nil, err
	}
	descReader := transform.NewReader(bytes.NewReader(descGbkData), simplifiedchinese.GBK.NewDecoder())
	descUtf8Html, err := io.ReadAll(descReader)
	if err != nil {
		return nil, err
	}
	descDoc, err := goquery.NewDocumentFromReader(bytes.NewReader(descUtf8Html))
	if err != nil {
		return nil, err
	}
	// 获取 JSON 文本
	descJsonStr := strings.TrimSpace(descDoc.Find("textarea.data-json-container").Text())
	var descData struct {
		Trustable bool `json:"trustable"`
		Sections  []struct {
			Title string `json:"title"`
			Value struct {
				Format  string `json:"format"`
				Content string `json:"content"`
			} `json:"value"`
		} `json:"sections"`
	}
	if err := json.Unmarshal([]byte(descJsonStr), &descData); err != nil {
		return nil, metaerror.Wrap(err, "parse description json failed")
	}

	description := ""

	for _, sec := range descData.Sections {
		title := strings.TrimSpace(sec.Title)
		htmlContent := sec.Value.Content
		if title == "" {
			md, err := foundationrender.HTMLToMarkdown(newProblemId, htmlContent, baseURL)
			if err != nil {
				return nil, metaerror.Wrap(err, "convert section HTML to markdown failed")
			}
			description += md + "\n\n"
		}
	}

	// 标题
	title := strings.TrimSpace(
		doc.Find("h2").Clone().Children().Remove().End().Text(),
	)
	jsonStr := strings.TrimSpace(doc.Find(`textarea[name="dataJson"]`).Text())

	var data struct {
		Properties []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"properties"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	var (
		timeLimit   int
		memoryLimit int
		source      string
	)

	for _, p := range data.Properties {
		switch p.Title {
		case "time_limit":
			// "1000 ms" -> 1000
			timeLimit, _ = strconv.Atoi(strings.Fields(p.Content)[0])
		case "mem_limit":
			// "65536 kB" -> 65536
			memoryLimit, _ = strconv.Atoi(strings.Fields(p.Content)[0])
		case "source":
			source = p.Content
		}
	}

	// 转成 goquery 文档片段
	srcDoc, err := goquery.NewDocumentFromReader(strings.NewReader(source))
	if err != nil {
		return nil, err
	}

	// 提取来源链接和文本
	a := srcDoc.Find("a").First()
	link, _ := a.Attr("href")
	sourceName := strings.TrimSpace(a.Text())

	// a 标签后面的文本中包含 author
	restText := a.Parent().Text()
	restText = strings.TrimSpace(restText)
	author := ""
	// 解析作者
	if idx := strings.Index(restText, "Author:"); idx >= 0 {
		author = strings.TrimSpace(restText[idx+len("Author:"):])
	}
	// 你最终存储的 source 可以按需求组合：
	source = fmt.Sprintf("[%s](%s)", sourceName, link)

	originUrl := fmt.Sprintf("http://poj.org/problem?id=%s", id)

	problem := foundationmodel.NewProblemBuilder().
		Title(title).
		Description(description).
		TimeLimit(timeLimit).
		MemoryLimit(memoryLimit).
		Source(&source).
		InsertTime(nowTime).
		ModifyTime(nowTime).
		Build()

	problemRemote := foundationmodel.NewProblemRemoteBuilder().
		OriginOj("POJ").
		OriginId(id).
		OriginUrl(originUrl).
		OriginAuthor(&author).
		Build()

	err = foundationdao.GetProblemDao().UpdateProblemCrawl(ctx, newProblemId, problem, problemRemote)
	if err != nil {
		return nil, err
	}

	return &newProblemId, nil
}

func (s *RemotePojAgent) PostCrawlProblem(ctx context.Context, id string) (*string, error) {

	err := s.login(ctx)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to login")
	}

	return s.crawlProblem(ctx, id, 0)
}

func (s *RemotePojAgent) cleanCodeBeforeSubmit(code string, language foundationjudge.JudgeLanguage) string {
	return code
}

func (s *RemotePojAgent) getMaxRunId(ctx context.Context, username string, problemId string) (string, error) {
	hduUrl := fmt.Sprintf(
		"https://acm.hdu.edu.cn/status.php?pid=%s&user=%s",
		problemId,
		username,
	)
	method := "GET"
	req, err := http.NewRequestWithContext(ctx, method, hduUrl, nil)
	if err != nil {
		return "", metaerror.Wrap(err, "failed to create getMaxRunId request")
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "acm.hdu.edu.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://acm.hdu.edu.cn/status.php?first=&pid=&user=&lang=0&status=0")
	res, err := s.goJudgeClient.Do(req)
	if err != nil {
		return "", metaerror.Wrap(err, "getMaxRunId request failed")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err))
		}
	}(res.Body)
	// 解析 HTML，获取最新的 Run ID
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", metaerror.Wrap(err, "failed to parse getMaxRunId response body")
	}
	var runId string
	doc.Find("div#fixed_table table").First().Find("tr").EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			firstTd := s.Find("td").First()
			if firstTd == nil {
				return true
			}
			text := strings.TrimSpace(firstTd.Text())
			if text == "" || text == "Run ID" {
				return true
			}
			runId = text
			return false
		},
	)
	return runId, nil
}

func (s *RemotePojAgent) requestJudgeJobStatus(ctx context.Context, runId string, retryCount int) (
	foundationjudge.JudgeStatus,
	int,
	int,
	int,
	error,
) {
	hduUrl := fmt.Sprintf("https://acm.hdu.edu.cn/status.php?first=%s", runId)
	method := "GET"
	req, err := http.NewRequestWithContext(ctx, method, hduUrl, nil)
	if err != nil {
		return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(
			err,
			"failed to create GetJudgeJobStatus request",
		)
	}
	req.Header.Add("Cookie", s.cookie)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "acm.hdu.edu.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", fmt.Sprintf("https://acm.hdu.edu.cn/status.php?first=&pid=&user=&lang=0&status=0"))
	res, err := s.goJudgeClient.Do(req)
	if err != nil {
		return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(err, "GetJudgeJobStatus request failed")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err))
		}
	}(res.Body)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(err, "failed to read response body")
	}
	bodyStr := string(body)
	if strings.Contains(bodyStr, "<title>User Login</title>") {
		if retryCount > 0 {
			return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.New("HDU remote judge login failed after retry")
		}
		// 重新登录
		err := s.login(ctx)
		if err != nil {
			return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(err, "failed to login")
		}
		return s.requestJudgeJobStatus(ctx, runId, retryCount+1)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyStr))
	if err != nil {
		return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(err, "failed to parse response body")
	}
	var statusStr, exeTimeStr, exeMemoryStr string
	doc.Find("div#fixed_table table").First().Find("tr").EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			tds := s.Find("td")
			if tds.Length() < 6 {
				return true
			}
			firstTdText := strings.TrimSpace(tds.Eq(0).Text())
			if firstTdText == "" || firstTdText == "Run ID" {
				return true
			}
			statusStr = strings.TrimSpace(tds.Eq(2).Text())
			exeTimeStr = strings.TrimSpace(tds.Eq(4).Text())
			exeMemoryStr = strings.TrimSpace(tds.Eq(5).Text())
			return false
		},
	)
	status := s.GetJudgeStatus(statusStr)
	score := 0
	exeTime := 0
	if strings.HasSuffix(exeTimeStr, "MS") {
		exeTimeStr = strings.TrimSuffix(exeTimeStr, "MS")
		exeTimeStr = strings.TrimSpace(exeTimeStr)
		exeTime, err = strconv.Atoi(exeTimeStr)
		if err != nil {
			return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(
				err,
				"failed to parse execution time",
			)
		}
		exeTime = exeTime * 1000000
	}
	exeMemory := 0
	if strings.HasSuffix(exeMemoryStr, "K") {
		exeMemoryStr = strings.TrimSuffix(exeMemoryStr, "K")
		exeMemoryStr = strings.TrimSpace(exeMemoryStr)
		exeMemory, err = strconv.Atoi(exeMemoryStr)
		if err != nil {
			return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(
				err,
				"failed to parse execution memory",
			)
		}
		exeMemory = exeMemory * 1024
	}
	if status == foundationjudge.JudgeStatusAC {
		score = 1000
	}
	return status, score, exeTime, exeMemory, nil
}

func (s *RemotePojAgent) GetJudgeJobStatus(ctx context.Context, id string) (
	foundationjudge.JudgeStatus,
	int,
	int,
	int,
	error,
) {
	return s.requestJudgeJobStatus(ctx, id, 0)
}

func (s *RemotePojAgent) GetJudgeJobExtraMessage(
	ctx context.Context,
	id string,
	status foundationjudge.JudgeStatus,
) (string, error) {
	if status != foundationjudge.JudgeStatusCE {
		return "", nil
	}
	hduUrl := fmt.Sprintf("https://acm.hdu.edu.cn/viewerror.php?rid=%s", id)
	method := "GET"
	req, err := http.NewRequestWithContext(ctx, method, hduUrl, nil)
	if err != nil {
		return "", metaerror.Wrap(err, "failed to create GetJudgeJobExtraMessage request")
	}
	req.Header.Add("Cookie", s.cookie)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "acm.hdu.edu.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", fmt.Sprintf("https://acm.hdu.edu.cn/status.php?first=&pid=&user=&lang=0&status=0"))
	res, err := s.goJudgeClient.Do(req)
	if err != nil {
		return "", metaerror.Wrap(err, "GetJudgeJobExtraMessage request failed")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err))
		}
	}(res.Body)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", metaerror.Wrap(err, "failed to parse GetJudgeJobExtraMessage response body")
	}
	var compileMessage string
	doc.Find("pre").EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			compileMessage = strings.TrimSpace(s.Text())
			return false
		},
	)
	return compileMessage, nil
}

func (s *RemotePojAgent) submit(
	ctx context.Context, problemId string,
	language foundationjudge.JudgeLanguage,
	code string, retryCount int,
) (string, string, error) {

	languageCode := s.getLanguageCode(language)
	if languageCode == "" {
		return "", "", metaerror.New("HDU remote judge not support language")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	hduUrl := "https://acm.hdu.edu.cn/submit.php?action=submit"
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	code = s.cleanCodeBeforeSubmit(code, language)

	urlEncoded := url.QueryEscape(code)
	base64Encoded := base64.StdEncoding.EncodeToString([]byte(urlEncoded))

	err := writer.WriteField("_usercode", base64Encoded)
	if err != nil {
		return "", "", metaerror.Wrap(err, "failed to write _usercode field")
	}
	err = writer.WriteField("problemid", problemId)
	if err != nil {
		return "", "", metaerror.Wrap(err, "failed to write problemid field")
	}
	err = writer.WriteField("language", languageCode)
	if err != nil {
		return "", "", metaerror.Wrap(err, "failed to write language field")
	}
	err = writer.Close()
	if err != nil {
		return "", "", metaerror.Wrap(err, "failed to close writer")
	}
	req, err := http.NewRequestWithContext(ctx, method, hduUrl, payload)
	if err != nil {
		return "", "", metaerror.Wrap(err, "failed to create request")
	}
	req.Header.Add("Cookie", s.cookie)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "acm.hdu.edu.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://acm.hdu.edu.cn/submit.php?action=submit")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := s.goJudgeClient.Do(req)
	if err != nil {
		return "", "", metaerror.Wrap(err, "failed to do request")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err))
		}
	}(res.Body)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", metaerror.Wrap(err, "failed to read response body")
	}
	bodyStr := string(body)
	if strings.Contains(bodyStr, "<title>User Login</title>") {
		if retryCount > 0 {
			return "", "", metaerror.New("HDU remote judge login failed after retry")
		}
		// 重新登录
		err := s.login(ctx)
		if err != nil {
			return "", "", metaerror.Wrap(err, "failed to login")
		}
		return s.submit(ctx, problemId, language, code, retryCount+1)
	}
	if !strings.Contains(bodyStr, "<title>Realtime Status</title>") {
		return "", "", metaerror.New("HDU remote judge submit failed")
	}
	username := foundationconfig.GetConfig().Remote.VJudge.Username
	runId, err := s.getMaxRunId(ctx, username, problemId)
	if err != nil {
		return "", "", metaerror.Wrap(err, "failed to get max run id")
	}
	return runId, username, nil
}

func (s *RemotePojAgent) PostSubmitJudgeJob(
	ctx context.Context,
	problemId string,
	language foundationjudge.JudgeLanguage,
	code string,
) (string, string, error) {
	if s.cookie == "" {
		err := s.login(ctx)
		if err != nil {
			return "", "", metaerror.Wrap(err, "failed to login")
		}
	}
	return s.submit(ctx, problemId, language, code, 0)
}
