package main

import (
	"flag"
	"fmt"
	"github.com/r3inbowari/clash/config"
	C "github.com/r3inbowari/clash/constant"
	"github.com/r3inbowari/clash/hub"
	"github.com/r3inbowari/clash/hub/executor"
	"github.com/r3inbowari/clash/log"
	"github.com/r3inbowari/common"
	. "github.com/r3inbowari/zlog"
	"go.uber.org/automaxprocs/maxprocs"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"
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

func main() {
	if runtime.GOOS == "windows" {
		_ = common.SetCmdTitle("Clash for Windows v1.9.0")
	}

	InitGlobalLogger().SetScreen(true)

	p := common.InitPermClient(common.PermOptions{
		Log:         &Log.Logger,
		CheckSource: "https://1077739472743245.cn-hangzhou.fc.aliyuncs.com/2016-08-15/proxy/perm.LATEST/perm",
		AppId:       "ef3d84021a", ExpireAfter: time.Hour * 168,
	})
	p.Verify()

	Log.Blue("   _     _      _     _      _     _      _     _      _     _   ")
	Log.Blue("  (c).-.(c)    (c).-.(c)    (c).-.(c)    (c).-.(c)    (c).-.(c)          PACKAGER #UNOFFICIAL " + "c4f7d8e...d93a5f2")
	Log.Blue("   / ._. \\      / ._. \\      / ._. \\      / ._. \\      / ._. \\            -... .. .-.. .. -.-. --- .. -. v1.9.0")
	Log.Blue(" __\\( Y )/__  __\\( Y )/__  __\\( Y )/__  __\\( Y )/__  __\\( Y )/__         Running: CLI Server" + " by cyt(r3inbowari)")
	Log.Blue("(_.-/'-'\\-._)(_.-/'-'\\-._)(_.-/'-'\\-._)(_.-/'-'\\-._)(_.-/'-'\\-._)        Listened: 6564")
	Log.Blue("   || C ||      || L ||      || A ||      || S ||      || H ||           PID: " + strconv.Itoa(os.Getpid()))
	Log.Blue(" _.' `-' '._  _.' `-' '._  _.' `-' '._  _.' `-' '._  _.' `-' '._         Built at: 2022.01.08 20:19:27")
	Log.Blue("(.-./`-'\\.-.)(.-./`-'\\.-.)(.-./`-'\\.-.)(.-./`-`\\.-.)(.-./`-'\\.-.)")
	Log.Blue(" `-'     `-'  `-'     `-'  `-'     `-'  `-'     `-'  `-'     `-' ")

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
