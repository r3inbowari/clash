package main

import (
	"crypto/md5"
	"encoding/hex"
	. "github.com/klauspost/cpuid/v2"
	. "github.com/r3inbowari/zlog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func ConfirmPermissions() {
	id := GetID()
	if Defs.PDigests == nil {
		Log.Error("[SYS] permissions list fetch error")
		return
	}
	for _, v := range Defs.PDigests {
		if id == v {
			Log.Infof("[SYS] permissions key -> %s [verified]", id)
			return
		}
	}
	Log.Info("[SYS] permissions key -> " + id)
	ExitOops()
}

func ExitOops() {
	Log.Warn("[SYS] oops, you don't have permission. please contact the developer (⑉･̆-･̆⑉)")
	time.Sleep(time.Minute * 3)
	os.Exit(999)
}

func GetID() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("[SYS] panic" + err.Error())
	}
	add := ""
	for _, inter := range interfaces {
		if inter.Name == "以太网" {
			add += inter.HardwareAddr.String()
		}
	}
	add += MixCPUInfo()
	return GetMD5(add)
}

func MixCPUInfo() string {
	v := strings.Join(CPU.FeatureSet(), ",")
	r := CPU.PhysicalCores + CPU.ThreadsPerCore + CPU.LogicalCores + CPU.Family + CPU.Model + CPU.CacheLine + CPU.Cache.L1D*CPU.Cache.L1I
	t := strconv.Itoa(r<<3 + int(CPU.VendorID))
	return GetMD5(v + t)
}

// GetMD5 md5
func GetMD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
