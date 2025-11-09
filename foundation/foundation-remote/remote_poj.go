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
	"log/slog"
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
	newProblemId := fmt.Sprintf("POJ-%s", id)

	baseURL := "http://poj.org/"
	problemUrl := fmt.Sprintf("%sproblem?id=%s", baseURL, id)

	resp, err := http.Get(problemUrl)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch problem page")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err))
		}
	}(resp.Body)

	utf8Html, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	utf8String := string(utf8Html)

	if strings.Contains(utf8String, "<li>Can not find problem (ID:") {
		return nil, nil
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(utf8String))
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(doc.Find(".ptt").First().Text())

	sectionMap := make(map[string]string)
	doc.Find("p.pst").Each(
		func(_ int, s *goquery.Selection) {
			key := strings.TrimSpace(s.Text())
			next := s.Next()

			if goquery.NodeName(next) == "div" && next.HasClass("ptx") {
				htmlContent, _ := next.Html()
				htmlContent, err = foundationrender.HTMLToMarkdown(newProblemId, htmlContent, baseURL)
				if err == nil {
					sectionMap[key] = htmlContent
				}
				return
			}

			// <pre class="sio"> 样例
			if goquery.NodeName(next) == "pre" && next.HasClass("sio") {
				sectionMap[key] = fmt.Sprintf("```\n%s\n```", next.Text())
				return
			}
		},
	)

	description := sectionMap["Description"]
	input := sectionMap["Input"]
	output := sectionMap["Output"]
	sampleInput := sectionMap["Sample Input"]
	sampleOutput := sectionMap["Sample Output"]

	hint := sectionMap["Hint"]
	if hint != "" {
		hint = "\n\n## Hint\n\n" + hint
	}

	source := sectionMap["Source"]

	limitText := doc.Find(".plm").Text()

	reTL := regexp.MustCompile(`Time Limit:\s*(\d+)MS`)
	reML := regexp.MustCompile(`Memory Limit:\s*(\d+)K`)

	timeLimit := -1
	memoryLimit := -1

	if m := reTL.FindStringSubmatch(limitText); len(m) > 1 {
		timeLimit, _ = strconv.Atoi(m[1])
	}
	if m := reML.FindStringSubmatch(limitText); len(m) > 1 {
		memoryLimit, _ = strconv.Atoi(m[1])
	}
	template := config.GetOjTemplateContent("poj")
	descriptionRendered := foundationrender.Render(
		template, map[string]string{
			"description":  description,
			"input":        input,
			"output":       output,
			"sampleInput":  sampleInput,
			"sampleOutput": sampleOutput,
			"hint":         hint,
		},
	)

	originUrl := fmt.Sprintf("http://poj.org/problem?id=%s", id)
	originAuthor := ""

	problem := foundationmodel.NewProblemBuilder().
		Title(title).
		Description(descriptionRendered).
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
		OriginAuthor(&originAuthor).
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

	slog.Info("POJ remote judge login start")

	loginUrl := "https://acm.poj.edu.cn/userloginex.php?action=login"
	method := "POST"
	username := foundationconfig.GetConfig().Remote.Poj.Username
	password := foundationconfig.GetConfig().Remote.Poj.Password

	payload := strings.NewReader(fmt.Sprintf("username=%s&userpass=%s", username, password))
	req, err := http.NewRequestWithContext(ctx, method, loginUrl, payload)
	if err != nil {
		return metaerror.Wrap(err, "failed to create login request")
	}
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "acm.poj.edu.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://acm.poj.edu.cn/userloginex.php?action=login")
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
	pojUrl := fmt.Sprintf(
		"https://acm.poj.edu.cn/status.php?pid=%s&user=%s",
		problemId,
		username,
	)
	method := "GET"
	req, err := http.NewRequestWithContext(ctx, method, pojUrl, nil)
	if err != nil {
		return "", metaerror.Wrap(err, "failed to create getMaxRunId request")
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "acm.poj.edu.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://acm.poj.edu.cn/status.php?first=&pid=&user=&lang=0&status=0")
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
	pojUrl := fmt.Sprintf("https://acm.poj.edu.cn/status.php?first=%s", runId)
	method := "GET"
	req, err := http.NewRequestWithContext(ctx, method, pojUrl, nil)
	if err != nil {
		return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(
			err,
			"failed to create GetJudgeJobStatus request",
		)
	}
	req.Header.Add("Cookie", s.cookie)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "acm.poj.edu.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", fmt.Sprintf("https://acm.poj.edu.cn/status.php?first=&pid=&user=&lang=0&status=0"))
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
			return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.New("POJ remote judge login failed after retry")
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
	pojUrl := fmt.Sprintf("https://acm.poj.edu.cn/viewerror.php?rid=%s", id)
	method := "GET"
	req, err := http.NewRequestWithContext(ctx, method, pojUrl, nil)
	if err != nil {
		return "", metaerror.Wrap(err, "failed to create GetJudgeJobExtraMessage request")
	}
	req.Header.Add("Cookie", s.cookie)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "acm.poj.edu.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", fmt.Sprintf("https://acm.poj.edu.cn/status.php?first=&pid=&user=&lang=0&status=0"))
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

func (s *RemotePojAgent) submitImp(
	ctx context.Context, problemId string,
	language foundationjudge.JudgeLanguage,
	code string, retryCount int,
) (string, string, error) {

	languageCode := s.getLanguageCode(language)
	if languageCode == "" {
		return "", "", metaerror.New("POJ remote judge not support language")
	}

	pojUrl := "https://acm.poj.edu.cn/submit.php?action=submit"
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
	req, err := http.NewRequestWithContext(ctx, method, pojUrl, payload)
	if err != nil {
		return "", "", metaerror.Wrap(err, "failed to create request")
	}
	req.Header.Add("Cookie", s.cookie)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "acm.poj.edu.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://acm.poj.edu.cn/submit.php?action=submit")
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
			return "", "", metaerror.New("POJ remote judge login failed after retry")
		}
		// 重新登录
		err := s.login(ctx)
		if err != nil {
			return "", "", metaerror.Wrap(err, "failed to login")
		}
		return s.submitImp(ctx, problemId, language, code, retryCount+1)
	}
	if !strings.Contains(bodyStr, "<title>Realtime Status</title>") {
		return "", "", metaerror.New("POJ remote judge submit failed")
	}
	username := foundationconfig.GetConfig().Remote.Poj.Username
	runId, err := s.getMaxRunId(ctx, username, problemId)
	if err != nil {
		return "", "", metaerror.Wrap(err, "failed to get max run id")
	}
	return runId, username, nil
}

func (s *RemotePojAgent) submit(
	ctx context.Context, problemId string,
	language foundationjudge.JudgeLanguage,
	code string, retryCount int,
) (string, string, error) {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	slog.Info("POJ remote judge submit start", "problemId", problemId, "language", language)

	return s.submitImp(ctx, problemId, language, code, retryCount)
}

func (s *RemotePojAgent) PostSubmitJudgeJob(
	ctx context.Context,
	problemId string,
	language foundationjudge.JudgeLanguage,
	code string,
) (string, string, error) {
	//if s.cookie == "" {
	//	err := s.login(ctx)
	//	if err != nil {
	//		return "", "", metaerror.Wrap(err, "failed to login")
	//	}
	//}
	//return s.submit(ctx, problemId, language, code, 0)
	return "", "", metaerror.New("POJ remote judge is disabled")
}
