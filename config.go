/**
 *
 * @author  chosen0ne(louzhenlin86@126.com)
 * @date    2017-03-15 11:22:12
 */

package redisproxy

import (
	"bytes"
	"fmt"
	"github.com/chosen0ne/goconf"
)

type redisConfig struct {
	Port            int
	MaxClients      int
	LogPath         string
	LogFile         string
	LogLevel        string
	CliMaxIdleTime  int // in seconds
	FromCliChanSize int
	EnableDebug     bool
}

func (config *redisConfig) String() string {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, "\tredisConifg: {\r\n")
	fmt.Fprintf(buf, "\tport: %d\r\n", config.Port)
	fmt.Fprintf(buf, "\tmax-clients: %d\r\n", config.MaxClients)
	fmt.Fprintf(buf, "\tlog-path: %s\r\n", config.LogPath)
	fmt.Fprintf(buf, "\tlog-file: %s\r\n", config.LogFile)
	fmt.Fprintf(buf, "\tlog-level: %s\r\n", config.LogLevel)
	fmt.Fprintf(buf, "\tcli-max-idle-time: %d\r\n", config.CliMaxIdleTime)
	fmt.Fprintf(buf, "\tfrom-cli-chan-size: %d\r\n", config.FromCliChanSize)
	fmt.Fprintf(buf, "\tenabel-debug: %t\r\n", config.EnableDebug)
	buf.WriteString("}")

	return string(buf.Bytes())
}

func (srv *redisServer) loadConfig(confPath string) (*redisConfig, error) {
	conf := &redisConfig{}
	setDefaultConfig(conf)

	if err := goconf.Load(conf, confPath); err != nil {
		return nil, err
	}

	return conf, nil
}

func setDefaultConfig(conf *redisConfig) {
	conf.Port = 8651
	conf.MaxClients = 10000
	conf.LogPath = "./"
	conf.LogFile = "redis.log"
	conf.LogLevel = "INFO"
	conf.CliMaxIdleTime = 15
	conf.FromCliChanSize = 100
}
