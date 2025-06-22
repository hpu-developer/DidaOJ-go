package foundationjudge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	gojudge "judge/go-judge"
	"log/slog"
	metaerror "meta/meta-error"
	metapanic "meta/meta-panic"
	metastring "meta/meta-string"
	"meta/retry"
	"net/http"
	"time"
)

func CompileCode(
	client *http.Client,
	jobKey string,
	runUrl string,
	language JudgeLanguage,
	code string,
	configFiles map[string]string,
) (map[string]string, string, JudgeStatus, error) {
	slog.Info("compile code", "job", jobKey)

	var args []string
	var copyIns map[string]interface{}
	var copyOutCached []string

	cpuLimit := 10000000000      // 10秒
	memoryLimit := 1048576 * 256 // 256MB

	env := []string{"PATH=/usr/bin:/usr/local/bin:/bin"}

	switch language {
	case JudgeLanguageC:
		args = []string{
			"gcc",
			"-fno-asm",
			"-fmax-errors=10",
			"-Wall",
			"--static",
			"-DONLINE_JUDGE",
			"-o",
			"a",
			"a.c",
			"-lm",
		}
		copyIns = map[string]interface{}{
			"a.c": map[string]interface{}{
				"content": code,
			},
		}
		copyOutCached = []string{"a"}
		break
	case JudgeLanguageCpp:
		args = []string{
			"g++", "-fno-asm", "-fmax-errors=10", "-Wall", "--static",
			"-DONLINE_JUDGE", "-Wno-sign-compare",
			"-o", "a", "a.cc",
		}
		copyIns = map[string]interface{}{
			"a.cc": map[string]interface{}{
				"content": code,
			},
		}
		copyOutCached = []string{"a"}
		break
	case JudgeLanguageJava:
		cmd := "javac -J-Xms128m -J-Xmx512m -encoding UTF-8 -Xlint:unchecked Main.java && jar -cvf Main.jar *.class"
		args = []string{"bash", "-c", cmd}
		copyIns = map[string]interface{}{
			"Main.java": map[string]interface{}{
				"content": code,
			},
		}
		copyOutCached = []string{"Main.jar"}
		break
	case JudgeLanguagePython:
		args = []string{"python3", "-c", "import py_compile; py_compile.compile(r'a.py')"}
		copyIns = map[string]interface{}{
			"a.py": map[string]interface{}{
				"content": code,
			},
		}
		copyOutCached = nil
	case JudgeLanguagePascal:
		args = []string{"fpc", "-Fu/usr/lib/x86_64-linux-gnu/fpc/3.2.2/units/x86_64-linux/rtl", "a.pas"}
		copyIns = map[string]interface{}{
			"a.pas": map[string]interface{}{
				"content": code,
			},
		}
		copyOutCached = []string{"a"}
	case JudgeLanguageGolang:
		env = append(env, "GOCACHE=/tmp/go_cache")
		args = []string{"go", "build", "-o", "a"}
		copyIns = map[string]interface{}{
			"a.go": map[string]interface{}{
				"content": code,
			},
			"go.mod": map[string]interface{}{
				"content": "module main\n",
			},
		}
		copyOutCached = []string{"a"}
	case JudgeLanguageTypeScript:
		args = []string{"bash", "-c", "tar -xzf ts-env.tar.gz && npx tsc"}
		env = append(env, "HOME=/tmp/judge")
		copyIns = map[string]interface{}{
			"a.ts": map[string]interface{}{
				"content": code,
			},
			"ts-env.tar.gz": map[string]interface{}{
				"fileId": configFiles["ts-env"],
			},
		}
		copyOutCached = []string{"a.js"}
		cpuLimit = cpuLimit + 5000000000        // TypeScript 编译时间可能较长，增加5秒
		memoryLimit = memoryLimit + 1048576*128 // TypeScript 编译可能需要更多内存，增加128MB
	default:
		return nil, "compile failed, language not support.",
			JudgeStatusJudgeFail,
			metaerror.New("language not support: %d", language)
	}

	// 准备请求数据
	data := map[string]interface{}{
		"cmd": []map[string]interface{}{
			{
				"args": args,
				"env":  env,
				"files": []map[string]interface{}{
					{"content": ""},
					{"name": "stdout", "max": 10240},
					{"name": "stderr", "max": 10240},
				},
				"cpuLimit":      cpuLimit,    // 10秒
				"memoryLimit":   memoryLimit, // 256MB
				"procLimit":     50,
				"copyIn":        copyIns,
				"copyOut?":      []string{"stdout", "stderr"},
				"copyOutCached": copyOutCached,
			},
		},
	}

	var finalMessage string
	var finalErr error
	var responseDataList []struct {
		Status gojudge.Status `json:"status"`
		Error  string         `json:"error"`
		Files  struct {
			Stderr string `json:"stderr"`
			Stdout string `json:"stdout"`
		} `json:"files"`
		FileIds   map[string]string `json:"fileIds"`
		FileError []struct {
			Name    string `json:"name"`
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"fileError"`
	}
	_ = retry.TryRetrySleep(
		"compile_code", 3, time.Second*3, func(int) bool {
			jsonData, err := json.Marshal(data)
			if err != nil {
				finalMessage = "compile failed, request data marshal error."
				finalErr = metaerror.Wrap(err, "failed to marshal request data")
				return true
			}
			request, err := http.NewRequest(http.MethodPost, runUrl, bytes.NewBuffer(jsonData))
			if err != nil {
				finalMessage = "compile failed, request data create error."
				finalErr = metaerror.Wrap(err, "failed to create request")
				return true
			}
			request.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(request)
			if err != nil {
				finalMessage = "compile failed, upload file error."
				finalErr = metaerror.Wrap(err, "failed to post request")
				return false
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					metapanic.ProcessError(err)
				}
			}(resp.Body)
			if resp.StatusCode != http.StatusOK {
				finalMessage = fmt.Sprintf("compile failed, response status code %d error.", resp.StatusCode)
				finalErr = metaerror.New(
					"unexpected status code: %d",
					resp.StatusCode,
				)
				return false
			}
			err = json.NewDecoder(resp.Body).Decode(&responseDataList)
			if err != nil {
				finalMessage = "compile failed, upload file response parse error."
				finalErr = metaerror.Wrap(err, "failed to decode response")
				return false
			}
			if len(responseDataList) != 1 {
				finalMessage = "compile failed, compile response data error."
				finalErr = metaerror.New(
					"unexpected response length: %d",
					len(responseDataList),
				)
				return false
			}
			finalMessage = ""
			finalErr = nil
			return true
		},
	)

	if finalErr != nil {
		return nil, finalMessage, JudgeStatusJudgeFail, finalErr
	}

	responseData := responseDataList[0]
	errorMessage := responseData.Error
	if responseData.Files.Stderr != "" {
		if errorMessage != "" {
			errorMessage += "\n"
		}
		errorMessage += responseData.Files.Stderr
	}
	if responseData.Files.Stdout != "" {
		if errorMessage != "" {
			errorMessage += "\n"
		}
		errorMessage += responseData.Files.Stdout
	}
	if errorMessage == "" {
		if len(responseData.FileError) > 0 {
			for _, fileError := range responseData.FileError {
				errorMessage += fmt.Sprintf(
					"File: %s, Type: %s, Message: %s\n",
					fileError.Name,
					fileError.Type,
					fileError.Message,
				)
			}
		}
	}
	errorMessage = metastring.GetTextEllipsis(errorMessage, 1000)
	if responseData.Status != gojudge.StatusAccepted {
		if responseData.Status != gojudge.StatusNonzeroExit &&
			responseData.Status != gojudge.StatusFileError {
			slog.Warn("compile error", "job", jobKey, "responseData", responseData)
			return nil, errorMessage, JudgeStatusCLE, nil
		} else {
			return nil, errorMessage, JudgeStatusCE, nil
		}
	}
	return responseData.FileIds, errorMessage, JudgeStatusAC, nil
}

func DeleteFile(jobKey string, deleteFileUrl string) error {
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodDelete, deleteFileUrl, nil)
	if err != nil {
		return err
	}
	_, err = client.Do(request)
	if err != nil {
		return metaerror.Wrap(err, "failed to delete file")
	}
	return nil
}
