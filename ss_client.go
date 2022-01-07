package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"io"
	"net"

	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/listener/socks"
)

type VPNOptions struct {
	outbound.ShadowSocksOption
	Addr string
	Log  *logrus.Logger
}

type SS struct {
	VPNOptions
	inbound  chan constant.ConnContext
	listener *socks.Listener

	ss *outbound.ShadowSocks
}

func InitSSClient(opts VPNOptions) (*SS, error) {

	var vpn SS
	var err error
	if opts.Log == nil {
		opts.Log = logrus.New()
	}

	vpn.VPNOptions = opts

	vpn.inbound = make(chan constant.ConnContext, 100)

	vpn.listener, err = socks.New(opts.Addr, vpn.inbound)
	if err != nil {
		return nil, err
	}

	vpn.ss, err = outbound.NewShadowSocks(opts.ShadowSocksOption)
	if err != nil {
		return nil, err
	}
	return &vpn, nil
}

func (v *SS) Start() {
	v.Log.Infof("[SS] listening on %s", v.listener.Address())
	for c := range v.inbound {
		conn := c
		metadata := conn.Metadata()
		v.Log.Infof("[SS] request incoming from %s to %s", metadata.SourceAddress(), metadata.RemoteAddress())
		go func() {
			remote, err := v.ss.DialContext(context.Background(), metadata)
			if err != nil {
				v.Log.Errorf("[SS] dial error: %s", err.Error())
				return
			}
			relay(remote, conn.Conn())
		}()
	}
}

func (v *SS) Close() {
	defer close(v.inbound)
	if v.listener != nil {
		v.listener.Close()
	}
	v.Log.Info("[SS] shadowSocks has been closed")
}

func relay(l, r net.Conn) {
	go io.Copy(l, r)
	io.Copy(r, l)
}
