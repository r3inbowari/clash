package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	. "github.com/r3inbowari/zlog"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

func CreateInstallBatch(name string) {
	file, e := os.OpenFile("install.bat", os.O_CREATE|os.O_WRONLY, 0666)
	if e != nil {
		fmt.Println("failed")
		os.Exit(1004)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	writer.WriteString("taskkill /f /pid " + strconv.Itoa(os.Getpid()) + "\n")
	writer.WriteString("start \"" + name + "\" " + name + ".exe\n")
	writer.WriteString("exit\n")
	writer.Flush()
}

func ExecBatchFromWindows(path string) error {
	return exec.Command("cmd.exe", "/c", "start "+path+".bat").Start()
}

func Reload(path string) error {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		exec.Command("chmod", "777", path)
		path = "./" + path
	}
	// init接管
	cmd := exec.Command(path, "-a")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

func DigestVerify(path string, ver string, digestStr string) bool {
	if runtime.GOOS == "windows" {
		path += "_" + ver + ".exe"
	}
	file, err := os.Open(path)
	if err != nil {
		Log.Error("[UP] file not exist")
		return false
	}
	md5f := md5.New()
	_, err = io.Copy(md5f, file)
	if err != nil {
		Log.Error("[UP] file open error")
		return false
	}

	ok := digestStr == hex.EncodeToString(md5f.Sum([]byte("")))
	if ok {
		Log.WithFields(logrus.Fields{"digest": digestStr, "file": hex.EncodeToString(md5f.Sum([]byte("")))}).Info("[UP] file digest match")
	} else {
		Log.WithFields(logrus.Fields{"digest": digestStr, "file": hex.EncodeToString(md5f.Sum([]byte("")))}).Warn("[UP] file digest mismatch")
	}
	return ok
}

var host = "http://r3in.top:3000/"

// "https://cdn.jsdelivr.net/gh/r3inbowari/hbuilderx_cli@v1.0.16/meiwobuxing_darwin_amd64_v1.0.15"
var speedup = "https://cdn.jsdelivr.net/gh/r3inbowari/hbuilderx_cli@"

type Default struct {
	Name     string   `json:"name"`
	Major    int      `json:"major"`
	Minor    int      `json:"minor"`
	Patch    int      `json:"patch"`
	Types    []string `json:"types"`
	Digests  []string `json:"digests"`
	PDigests []string `json:"pDigests"` // hi, caicai
	Desc     string   `json:"desc"`
}

var Defs *Default
var checkUpdateUrl = "https://1077739472743245.cn-hangzhou.fc.aliyuncs.com/2016-08-15/proxy/reg.LATEST/meiwobuxing/default"

type CheckResult struct {
	Total   int     `json:"total"`
	Data    Default `json:"data"`
	Code    int     `json:"code"`
	Message string  `json:"msg"`
}

func CheckUpdate() (bool, string, string) {
	Log.Info("[UP] redirecting to ws://cn-hangzhou.aliyuncs.com")
	res, err := http.Get(checkUpdateUrl)
	if err != nil {
		return false, "", ""
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, "", ""
	}
	var result CheckResult
	var defs Default
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, "", ""
	}
	defs = result.Data
	Defs = &result.Data
	Log.Info("[UP] ask: HELLO 2.9 (2.9.1) 2021-08-11.2118.f127dd6")

	Log.WithFields(logrus.Fields{"major": Up.Major, "minor": Up.Minor, "patch": Up.Patch}).Info("[UP] current version")
	value := defs.Major<<24 + defs.Minor<<12 + defs.Patch<<0
	now := Up.Major<<24 + Up.Minor<<12 + Up.Patch<<0
	if now < int64(value) {
		Log.WithFields(logrus.Fields{"major": defs.Major, "minor": defs.Minor, "patch": defs.Patch}).Info("[UP] found new version")
		if defs.Desc != "" {
			Log.Info("[UP] " + defs.Desc)
		}
		for k, v := range defs.Types {
			if v == runtime.GOOS+"_"+runtime.GOARCH {
				return true, defs.Digests[k], "v" + strconv.FormatInt(int64(defs.Major), 10) + "." + strconv.FormatInt(int64(defs.Minor), 10) + "." + strconv.FormatInt(int64(defs.Patch), 10)
			}
		}
	} else {
		Log.Info("[UP] the current version is up to date...")
	}
	return false, "", ""
}

func Auth(auth bool) {
	CheckUpdate()
	ConfirmPermissions()
	//cm := cron.New()
	//spec := "30 */30 * * * ?"
	//_ = cm.AddFunc(spec, func() {
	//	ConfirmPermissions()
	//})
	//cm.Start()
}

type Update struct {
	Patch           int64  // 0
	Minor           int64  // 0
	Major           int64  // 1
	VersionStr      string // "v1.0.0"
	BuildMode       string // dev
	ReleaseTag      string // "cb0dc838e04e841f193f383e06e9d25a534c5809"
	RuntimeOS       string // win
	BuildTime       string // 2021
	SucceedCallback func() // succeed
	AppName         string // app dir
	RunPath         string // 开发环境目录
}

var Up *Update

// InitUpdate 更新器件初始化
func InitUpdate(buildTime, buildMode string, ver, hash string, major, minor, patch string, name string, callback func()) *Update {
	var retUpdate Update

	retUpdate.AppName = name
	retUpdate.BuildMode = buildMode
	retUpdate.BuildTime = buildTime
	retUpdate.VersionStr = ver
	retUpdate.ReleaseTag = hash
	retUpdate.Major, _ = strconv.ParseInt(major, 10, 64)
	retUpdate.Minor, _ = strconv.ParseInt(minor, 10, 64)
	retUpdate.Patch, _ = strconv.ParseInt(patch, 10, 64)
	retUpdate.RuntimeOS = runtime.GOOS
	if callback != nil {
		retUpdate.SucceedCallback = callback
	} else {
		retUpdate.SucceedCallback = succeed
	}
	Up = &retUpdate

	var err error
	if Up.BuildMode != "DEV" {
		retUpdate.RunPath, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			Log.Error("[UP] unknown panic")
			time.Sleep(time.Second * 5)
			os.Exit(1005)
		}
	} else {
		// 开发环境路径
		// MACOS
		// retUpdate.RunPath = "/Users/r3inb/Downloads/meiwobuxing"
		// Windows
		retUpdate.RunPath = "C:\\Users\\inven\\Desktop\\meiwobuxing"
	}
	return &retUpdate
}

func succeed() {
	// 更新后
	// 重启前
	// 善后工作处理
	// do after update
}
