package main

import (
	"github.com/r3inbowari/clash/adapter/outbound"
	"github.com/r3inbowari/clash/client"
	"github.com/r3inbowari/common"
	"io"
	"net"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	ss, err := client.InitSSClient(client.VPNOptions{
		ShadowSocksOption: outbound.ShadowSocksOption{
			BasicOption: outbound.BasicOption{},
			Name:        "home",
			Server:      "invenleey.oicp.net",
			Port:        26185,
			Password:    "159463",
			Cipher:      "chacha20-ietf-poly1305",
			UDP:         false,
		},
		Addr: "127.0.0.1:6666",
	})

	if err != nil {
		println(err.Error())
	}

	go ss.Start()
	time.Sleep(time.Second * 2)

	go func() {
		common.InitMySqlByProxy(common.DBProxyOptions{
			User:   "root",
			Pwd:    "15598870762",
			Url:    "192.168.5.237",
			Port:   "3306",
			Schema: "hello",
			Socks:  "127.0.0.1:6666",
		})
	}()

	go func() {
		var res common.TaobaoBody
		common.RequestJson(common.RequestOptions{Url: "https://api.m.taobao.com/rest/api3.do?api=mtop.common.getTimestamp", Client: *common.RegisterHttpProxy(common.HttpProxyOptions{
			Socks:     "127.0.0.1:6666",
			SocksAuth: nil,
		})}, &res)

		println(res.Data.T)
	}()

	time.Sleep(time.Second * 4)

	ss.Close()
}

func relay(l, r net.Conn) {
	go io.Copy(l, r)
	io.Copy(r, l)
}
