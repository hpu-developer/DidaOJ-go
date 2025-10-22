package foundationremote

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	foundationconfig "foundation/foundation-config"
	foundationdao "foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationrender "foundation/foundation-render"
	"io"
	metaerror "meta/meta-error"
	metapanic "meta/meta-panic"
	metatime "meta/meta-time"
	"meta/singleton"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"web/config"

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

func (s *RemotePojAgent) PostCrawlProblem(ctx context.Context, id string) (*string, error) {

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
	doc.Find(".panel_title, .panel_content").Each(
		func(i int, s *goquery.Selection) {
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
		},
	)
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
	description := foundationrender.Render(
		template, map[string]string{
			"description":  contentMap["Problem Description"],
			"input":        contentMap["Input"],
			"output":       contentMap["Output"],
			"sampleInput":  contentMap["Sample Input"],
			"sampleOutput": contentMap["Sample Output"],
			"hint":         hint,
		},
	)

	originUrl := fmt.Sprintf("https://acm.hdu.edu.cn/showproblem.php?pid=%s", id)

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
		OriginOj("HDU").
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

func (s *RemotePojAgent) cleanCodeBeforeSubmit(code string, language foundationjudge.JudgeLanguage) string {
	return code
}

func (s *RemotePojAgent) login(ctx context.Context) error {
	url := "https://acm.hdu.edu.cn/userloginex.php?action=login"
	method := "POST"
	username := foundationconfig.GetConfig().Remote.VJudge.Username
	password := foundationconfig.GetConfig().Remote.VJudge.Password

	payload := strings.NewReader(fmt.Sprintf("username=%s&userpass=%s", username, password))
	req, err := http.NewRequestWithContext(ctx, method, url, payload)
	if err != nil {
		return metaerror.Wrap(err, "failed to create login request")
	}
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "acm.hdu.edu.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://acm.hdu.edu.cn/userloginex.php?action=login")
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
	s.cookie = cookie
	return nil
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
