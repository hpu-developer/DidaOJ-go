package foundationremote

import (
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
	"net/http"
	"net/http/cookiejar"
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
			jar, _ := cookiejar.New(nil)
			s.goJudgeClient = &http.Client{
				Transport: &http.Transport{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 100,
					MaxConnsPerHost:     100,
					IdleConnTimeout:     90 * time.Second,
				},
				Timeout: 60 * time.Second, // 请求整体超时
				Jar:     jar,
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
		return "3"
	case foundationjudge.JudgeLanguageJava:
		return "2"
	default:
		return ""
	}
}

func (s *RemotePojAgent) GetJudgeStatus(status string) foundationjudge.JudgeStatus {
	switch status {
	case "Waiting":
		return foundationjudge.JudgeStatusQueuing
	case "Compiling":
		return foundationjudge.JudgeStatusCompiling
	case "Running & Judging":
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
	case "Compile Error":
		return foundationjudge.JudgeStatusCE
	case "System Error":
		return foundationjudge.JudgeStatusRE
	case "Validator Error":
		return foundationjudge.JudgeStatusRE
	default:
		slog.Warn("unknown POJ judge status", "status", status)
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

	loginUrl := "http://poj.org/login"
	method := "POST"
	username := foundationconfig.GetConfig().Remote.Poj.Username
	password := foundationconfig.GetConfig().Remote.Poj.Password

	payload := strings.NewReader(fmt.Sprintf("user_id1=%s&password1=%s&B1=login&url=.", username, password))
	req, err := http.NewRequestWithContext(ctx, method, loginUrl, payload)
	if err != nil {
		return metaerror.Wrap(err, "failed to create login request")
	}
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
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
	return nil
}

func (s *RemotePojAgent) getMaxRunId(ctx context.Context, username string, problemId string) (string, error) {
	pojUrl := fmt.Sprintf(
		"http://poj.org/status?problem_id=%s&user_id=%s",
		problemId,
		username,
	)
	req, err := http.NewRequestWithContext(ctx, "GET", pojUrl, nil)
	if err != nil {
		return "", metaerror.Wrap(err, "failed to create getMaxRunId request")
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "poj.org")
	req.Header.Add("Connection", "keep-alive")
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
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", metaerror.Wrap(err, "failed to parse getMaxRunId response body")
	}
	table := doc.Find("table.a").First()
	if table.Length() == 0 {
		return "", metaerror.New("no status table found")
	}
	firstRow := table.Find("tr").Eq(1)
	if firstRow.Length() == 0 {
		return "", metaerror.New("no submissions found")
	}
	runId := strings.TrimSpace(firstRow.Find("td").First().Text())
	if runId == "" {
		return "", metaerror.New("runId not found")
	}
	return runId, nil
}

func (s *RemotePojAgent) requestJudgeJobStatus(ctx context.Context, runId string, retryCount int) (
	foundationjudge.JudgeStatus,
	int,
	int,
	int,
	error,
) {
	pojUrl := fmt.Sprintf("http://poj.org/showsource?solution_id=%s", runId)
	req, err := http.NewRequestWithContext(ctx, "GET", pojUrl, nil)
	if err != nil {
		return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(
			err,
			"failed to create GetJudgeJobStatus request",
		)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "poj.org")
	req.Header.Add("Connection", "keep-alive")

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
	if strings.Contains(bodyStr, "<li>Source request declined.</li>") {
		if retryCount > 0 {
			return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.New("POJ remote judge login failed after retry")
		}
		// 重新登录
		if err := s.login(ctx); err != nil {
			return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(err, "failed to login")
		}
		return s.requestJudgeJobStatus(ctx, runId, retryCount+1)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyStr))
	if err != nil {
		return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(err, "failed to parse response body")
	}

	table := doc.Find("table").First()
	if table.Length() == 0 {
		return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.New("submission table not found")
	}

	// Result
	resultTd := table.Find("tr").Eq(2).Find("td").Eq(2) // 第三行第3列
	statusStr := strings.TrimSpace(resultTd.Text())
	statusStr = strings.TrimPrefix(statusStr, "Result: ")
	statusStr = strings.TrimSpace(statusStr)

	// Memory
	memoryTd := table.Find("tr").Eq(1).Find("td").Eq(0)                               // 第二行第1列
	exeMemoryStr := strings.TrimSpace(strings.TrimPrefix(memoryTd.Text(), "Memory:")) // "Memory: N/A" -> "N/A"

	// Time
	timeTd := table.Find("tr").Eq(1).Find("td").Eq(2)                           // 第二行第3列
	exeTimeStr := strings.TrimSpace(strings.TrimPrefix(timeTd.Text(), "Time:")) // "Time: N/A" -> "N/A"

	status := s.GetJudgeStatus(statusStr)

	score := 0
	exeTime := 0
	exeMemory := 0

	if exeTimeStr != "N/A" {
		exeTimeStr = strings.TrimSpace(strings.TrimSuffix(exeTimeStr, "MS"))
		if exeTimeStr != "" {
			exeTime, err = strconv.Atoi(exeTimeStr)
			if err != nil {
				return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(
					err,
					"failed to parse execution time",
				)
			}
			exeTime *= 1000000
		}
	}

	if exeMemoryStr != "N/A" {
		exeMemoryStr = strings.TrimSpace(strings.TrimSuffix(exeMemoryStr, "K"))
		if exeMemoryStr != "" {
			exeMemory, err = strconv.Atoi(exeMemoryStr)
			if err != nil {
				return foundationjudge.JudgeStatusJudgeFail, 0, 0, 0, metaerror.Wrap(
					err,
					"failed to parse execution memory",
				)
			}
			exeMemory *= 1024
		}
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
	pojUrl := fmt.Sprintf("http://poj.org/showcompileinfo?solution_id=%s", id)
	method := "GET"
	req, err := http.NewRequestWithContext(ctx, method, pojUrl, nil)
	if err != nil {
		return "", metaerror.Wrap(err, "failed to create GetJudgeJobExtraMessage request")
	}
	req.Header.Add("Host", "poj.org")
	req.Header.Add("Connection", "keep-alive")
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

	slog.Info("POJ remote judge submit start", "problemId", problemId, "language", language)

	languageCode := s.getLanguageCode(language)
	if languageCode == "" {
		return "", "", metaerror.New("POJ remote judge not support language")
	}

	pojUrl := "http://poj.org/submit"
	method := "POST"

	code = s.cleanCodeBeforeSubmit(code, language)

	base64Encoded := base64.StdEncoding.EncodeToString([]byte(code))

	data := url.Values{}
	data.Set("problem_id", problemId)
	data.Set("language", languageCode)
	data.Set("source", base64Encoded)
	data.Set("submit", "Submit")
	data.Set("encoded", "1")

	payload := strings.NewReader(data.Encode())

	req, err := http.NewRequestWithContext(ctx, method, pojUrl, payload)
	if err != nil {
		return "", "", metaerror.Wrap(err, "failed to create request")
	}
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
		return s.submit(ctx, problemId, language, code, retryCount+1)
	}
	if !strings.Contains(bodyStr, "Problem Status List</font>") {
		return "", "", metaerror.New("POJ remote judge submit failed")
	}
	username := foundationconfig.GetConfig().Remote.Poj.Username
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

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.cookie == "" {
		err := s.login(ctx)
		if err != nil {
			return "", "", metaerror.Wrap(err, "failed to login")
		}
	}

	return s.submit(ctx, problemId, language, code, 0)
}
