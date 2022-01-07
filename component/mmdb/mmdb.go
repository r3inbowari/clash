package mmdb

import (
	"sync"

	C "github.com/r3inbowari/clash/constant"
	"github.com/r3inbowari/clash/log"
	"github.com/oschwald/geoip2-golang"
	. "github.com/r3inbowari/zlog"
)

var (
	mmdb *geoip2.Reader
	once sync.Once
)

func LoadFromBytes(buffer []byte) {
	once.Do(func() {
		var err error
		mmdb, err = geoip2.FromBytes(buffer)
		if err != nil {
			log.Fatalln("Can't load mmdb: %s", err.Error())
		}
	})
}

func Verify() bool {
	instance, err := geoip2.Open(C.Path.MMDB())
	Log.WithTag("MMDB").WithField("epoch", instance.Metadata().BuildEpoch).Info("mmdb verify")
	if err == nil {
		instance.Close()
	}
	return err == nil
}

func Instance() *geoip2.Reader {
	once.Do(func() {
		var err error
		mmdb, err = geoip2.Open(C.Path.MMDB())
		if err != nil {
			log.Fatalln("Can't load mmdb: %s", err.Error())
		}
	})

	return mmdb
}
