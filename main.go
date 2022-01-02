package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Dreamacro/clash/config"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/hub"
	"github.com/Dreamacro/clash/hub/executor"
	"github.com/Dreamacro/clash/log"
	. "github.com/r3inbowari/zlog"
	"go.uber.org/automaxprocs/maxprocs"
	"golang.org/x/sys/windows/registry"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

var (
	flagset            map[string]bool
	version            bool
	testConfig         bool
	homeDir            string
	configFile         string
	externalUI         string
	externalController string
	secret             string
)

func init() {
	flag.StringVar(&homeDir, "d", "", "set configuration directory")
	flag.StringVar(&configFile, "f", "", "specify configuration file")
	flag.StringVar(&externalUI, "ext-ui", "", "override external ui directory")
	flag.StringVar(&externalController, "ext-ctl", "", "override external controller address")
	flag.StringVar(&secret, "secret", "", "override secret for RESTful API")
	flag.BoolVar(&version, "v", false, "show current version of clash")
	flag.BoolVar(&testConfig, "t", false, "test configuration and exit")
	flag.Parse()

	flagset = map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		flagset[f.Name] = true
	})
}

type Result struct {
	API  string   `json:"api"`
	V    string   `json:"v"`
	Ret  []string `json:"ret"`
	Data Data     `json:"data"`
}

type Data struct {
	T string `json:"t"`
}

func getTime() *Result {
	url := "http://api.m.taobao.com/rest/api3.do?api=mtop.common.getTimestamp"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var ret Result
	err = json.Unmarshal(body, &ret)
	if err != nil {
		println(err.Error())
		return nil
	}
	return &ret
}

func setTitle(title string) {
	kernel32, _ := syscall.LoadLibrary(`kernel32.dll`)
	sct, _ := syscall.GetProcAddress(kernel32, `SetConsoleTitleW`)
	syscall.Syscall(sct, 1, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))), 0, 0)
	syscall.FreeLibrary(kernel32)
}

func main() {
	setTitle("Clash for Windows v1.8.0")

	InitUpdate("2021.12.20 22:10:13", "server", "v1.8.0", "cb0dc838e04e841f193f383e06e9d25a534c5809", "1", "8", "0", "meiwobuxing", nil)
	InitGlobalLogger().SetScreen(true)
	time.Sleep(time.Second)

	Log.Blue("   _     _      _     _      _     _      _     _      _     _   ")
	Log.Blue("  (c).-.(c)    (c).-.(c)    (c).-.(c)    (c).-.(c)    (c).-.(c)          PACKAGER #UNOFFICIAL " + Up.ReleaseTag[:7] + "..." + Up.ReleaseTag[33:])
	Log.Blue("   / ._. \\      / ._. \\      / ._. \\      / ._. \\      / ._. \\            -... .. .-.. .. -.-. --- .. -. " + Up.VersionStr)
	Log.Blue(" __\\( Y )/__  __\\( Y )/__  __\\( Y )/__  __\\( Y )/__  __\\( Y )/__         Running: CLI Server" + " by cyt(r3inbowari)")
	Log.Blue("(_.-/'-'\\-._)(_.-/'-'\\-._)(_.-/'-'\\-._)(_.-/'-'\\-._)(_.-/'-'\\-._)        Listened: 6564")
	Log.Blue("   || C ||      || L ||      || A ||      || S ||      || H ||           PID: " + strconv.Itoa(os.Getpid()))
	Log.Blue(" _.' `-' '._  _.' `-' '._  _.' `-' '._  _.' `-' '._  _.' `-' '._         Built: " + Up.BuildTime)
	Log.Blue("(.-./`-'\\.-.)(.-./`-'\\.-.)(.-./`-'\\.-.)(.-./`-`\\.-.)(.-./`-'\\.-.)")
	Log.Blue(" `-'     `-'  `-'     `-'  `-'     `-'  `-'     `-'  `-'     `-' ")

	// Auth(true)

	t := getTime()
	if t == nil {
		return
	}

	it, err := strconv.ParseInt(t.Data.T, 10, 64)
	if err != nil {
		return
	}

	//if it > 1641312000000 {
	if it > 1640429981000 {
		Log.WithTag("[HUB]").Warn("Clash 授权已到期")
		time.Sleep(time.Minute)
		return
	}

	maxprocs.Set(maxprocs.Logger(func(string, ...interface{}) {}))
	if version {
		fmt.Printf("Clash %s %s %s with %s %s\n", C.Version, runtime.GOOS, runtime.GOARCH, runtime.Version(), C.BuildTime)
		return
	}

	if homeDir != "" {
		if !filepath.IsAbs(homeDir) {
			currentDir, _ := os.Getwd()
			homeDir = filepath.Join(currentDir, homeDir)
		}
		C.SetHomeDir(homeDir)
	}

	if configFile != "" {
		if !filepath.IsAbs(configFile) {
			currentDir, _ := os.Getwd()
			configFile = filepath.Join(currentDir, configFile)
		}
		C.SetConfig(configFile)
	} else {
		configFile := filepath.Join(C.Path.HomeDir(), C.Path.Config())
		C.SetConfig(configFile)
	}

	if err := config.Init(C.Path.HomeDir()); err != nil {
		log.Fatalln("Initial configuration directory error: %s", err.Error())
	}

	if testConfig {
		if _, err := executor.Parse(); err != nil {
			log.Errorln(err.Error())
			fmt.Printf("configuration file %s test failed\n", C.Path.Config())
			os.Exit(1)
		}
		fmt.Printf("configuration file %s test is successful\n", C.Path.Config())
		return
	}

	var options []hub.Option
	if flagset["ext-ui"] {
		options = append(options, hub.WithExternalUI(externalUI))
	}
	if flagset["ext-ctl"] {
		options = append(options, hub.WithExternalController(externalController))
	}
	if flagset["secret"] {
		options = append(options, hub.WithSecret(secret))
	}

	if err := hub.Parse(options...); err != nil {
		log.Fatalln("Parse config error: %s", err.Error())
	}

	key, _, _ := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.ALL_ACCESS)
	key.SetDWordValue("ProxyEnable", 1)
	key.SetStringValue("ProxyOverride", "localhost;127.*;10.*;172.16.*;172.17.*;172.18.*;172.19.*;172.20.*;172.21.*;172.22.*;172.23.*;172.24.*;172.25.*;172.26.*;172.27.*;172.28.*;172.29.*;172.30.*;172.31.*;192.168.*;<local>")
	key.SetStringValue("ProxyServer", "127.0.0.1:7890")
	Log.WithTag("MAIN").Info("open global proxy on windows setting")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	defer key.Close()
	Log.WithTag("MAIN").Info("close global proxy on windows setting")
	time.Sleep(time.Second)
	key.SetDWordValue("ProxyEnable", 0)
}
