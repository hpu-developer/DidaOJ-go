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
	"net/http"
)

func CompileCode(jobKey string, runUrl string, language JudgeLanguage, code string, configFiles map[string]string) (map[string]string, string, JudgeStatus, error) {
	slog.Info("compile code", "job", jobKey)

	var args []string
	var copyIns map[string]interface{}
	var copyOutCached []string

	cpuLimit := 10000000000      // 10秒
	memoryLimit := 1048576 * 256 // 256MB

	env := []string{"PATH=/usr/bin:/usr/local/bin:/bin"}

	switch language {
	case JudgeLanguageC:
		args = []string{"gcc", "-fno-asm", "-fmax-errors=10", "-Wall", "--static", "-DONLINE_JUDGE", "-o", "a", "a.c", "-lm"}
		copyIns = map[string]interface{}{
			"a.c": map[string]interface{}{
				"content": code,
			},
		}
		copyOutCached = []string{"a"}
		break
	case JudgeLanguageCpp:
		args = []string{"g++", "-fno-asm", "-fmax-errors=10", "-Wall", "--static",
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
		args = []string{"javac", "-J-Xms128m", "-J-Xmx512m", "-encoding", "UTF-8", "Main.java"}
		copyIns = map[string]interface{}{
			"Main.java": map[string]interface{}{
				"content": code,
			},
		}
		copyOutCached = []string{"Main.class"}
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
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, "compile failed, system error.", JudgeStatusJudgeFail, err
	}
	resp, err := http.Post(runUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "compile failed, upload file error.", JudgeStatusJudgeFail, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, "compile failed, upload file response error.", JudgeStatusJudgeFail, metaerror.New("unexpected status code: %d", resp.StatusCode)
	}
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
	err = json.NewDecoder(resp.Body).Decode(&responseDataList)
	if err != nil {
		return nil, fmt.Sprintf("compile failed, upload file response parse error."), JudgeStatusJudgeFail, metaerror.Wrap(err, "failed to decode response")
	}
	if len(responseDataList) != 1 {
		return nil, "compile failed, compile response data error.", JudgeStatusJudgeFail, metaerror.New("unexpected response length: %d", len(responseDataList))
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
				errorMessage += fmt.Sprintf("File: %s, Type: %s, Message: %s\n", fileError.Name, fileError.Type, fileError.Message)
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
