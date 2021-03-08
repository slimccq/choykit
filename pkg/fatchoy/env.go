// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"encoding/json"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"

	"devpkg.work/choykit/pkg/protocol"
	"devpkg.work/choykit/pkg/x/dotenv"
	"devpkg.work/choykit/pkg/x/strutil"
	"github.com/go-sql-driver/mysql"
)

type Environ protocol.Environ

// Environ的字段对应的.env变量名
var envFieldMapping = map[string]string{
	"Env":                       "APP_ENV",
	"GameId":                    "APP_GAME_ID",
	"ChannelId":                 "APP_CHANNEL_ID",
	"ServerName":                "APP_SERVER_NAME",
	"ServerId":                  "APP_SERVER_ID",
	"AccessKey":                 "APP_ACCESS_KEY",
	"ServiceType":               "APP_SERVICE_TYPE",
	"ServiceIndex":              "APP_SERVICE_INDEX",
	"ServiceDependency":         "APP_SERVICE_DEPENDENCY",
	"WorkingDir":                "APP_WORKING_DIR",
	"PprofAddr":                 "APP_PPROF_ADDR",
	"ExecutorCapacity":          "RUNTIME_EXECUTOR_CAPACITY",
	"ContextInboundQueueSize":   "RUNTIME_CONTEXT_INBOUND_SIZE",
	"ContextOutboundQueueSize":  "RUNTIME_CONTEXT_OUTBOUND_SIZE",
	"EndpointOutboundQueueSize": "RUNTIME_ENDPOINT_OUTBOUND_SIZE",
	"NetEnableEncryption":       "NET_ENABLE_ENCRYPTION",
	"NetPublicKeyFile":          "NET_PUBLIC_KEY_FILE",
	"NetPrivateKeyFile":         "NET_PRIVATE_KEY_FILE",
	"NetRpcTimeoutInterval":     "NET_RPC_TIMEOUT_INTERVAL",
	"NetPeerPingInterval":       "NET_PEER_PING_INTERVAL",
	"NetPeerReadTimeout":        "NET_PEER_READ_TIMEOUT",
	"NetSessionReadTimeout":     "NET_SESSION_READ_TIMEOUT",
	"DbMysqlDsn":                "DB_MYSQL_DSN",
	"DbRedisAddr":               "DB_REDIS_ADDR",
}

func (e *Environ) IsProd() bool {
	return e.Env == "prod"
}

func (e *Environ) SetByOption(opt *Options) {
	if opt.WorkingDir != "" {
		e.WorkingDir = opt.WorkingDir
	}
	if opt.ServiceType != "" {
		e.ServiceType = opt.ServiceType
	}
	if opt.ServiceIndex > 0 {
		e.ServiceIndex = int32(opt.ServiceIndex)
	}
	if opt.ServiceDependency != "" {
		e.ServiceDependency = opt.ServiceDependency
	}
	if opt.EtcdAddress != "" {
		e.EtcdAddr = opt.EtcdAddress
	}
	if opt.EtcdKeySpace != "" {
		e.EtcdKeyspace = opt.EtcdKeySpace
	}
	if opt.EtcdLeaseTTL > 0 {
		e.EtcdLeaseTtl = int32(opt.EtcdLeaseTTL)
	}
	if opt.LogLevel != "" {
		e.LogLevel = opt.LogLevel
	}
	if opt.EnableSysLog {
		e.EnableSyslog = opt.EnableSysLog
	}
	if opt.SysLogParams != "" {
		e.SyslogParams = opt.SysLogParams
	}
}

func NewEnviron() *Environ {
	return &Environ{
		EtcdAddr:                  "127.0.0.1:2379",
		EtcdKeyspace:              "choyd",
		EtcdLeaseTtl:              5,
		LogLevel:                  "debug",
		ExecutorCapacity:          20000, //
		ContextInboundQueueSize:   60000, //
		ContextOutboundQueueSize:  8000,  //
		EndpointOutboundQueueSize: 1000,  //
		NetRpcTimeoutInterval:     60,    // 60s
		NetPeerPingInterval:       10,    // 10s
		NetPeerReadTimeout:        100,   // 100s
		NetSessionReadTimeout:     180,   // 180s
	}
}

// 加载环境变量
func LoadEnviron() *Environ {
	env := NewEnviron()
	rv := reflect.ValueOf(env).Elem()
	rType := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		var name = rType.Field(i).Name
		var key = envFieldMapping[name]
		if key == "" {
			continue
		}
		var value = strings.TrimSpace(dotenv.Get(key))
		if value == "" {
			continue // 值为空的时候不覆盖
		}
		strutil.ParseStringToValue(value, rv.Field(i))
	}
	// 加载网络接口地址
	env.NetInterfaces = ParseNetInterface(dotenv.Get("NET_INTERFACES"))
	return env
}

func (e Environ) String() string {
	data, _ := json.Marshal(&e)
	return string(data)
}

// 解析地址格式，对外地址@bind地址:端口，如example.com@0.0.0.0:9527
func ParseNetInterface(text string) []*protocol.InterfaceAddr {
	var result []*protocol.InterfaceAddr
	addrItems := strings.Split(text, ",")
	for _, addrText := range addrItems {
		host, port, err := net.SplitHostPort(addrText)
		if err != nil {
			log.Panicf("parse address %s: %v", addrText, err)
		}
		n, _ := strconv.Atoi(port)
		addr := &protocol.InterfaceAddr{Port: int32(n)}
		i := strings.Index(host, "@")
		if i < 0 {
			addr.AdvertiseAddr = host
			addr.BindAddr = host
		} else {
			addr.AdvertiseAddr = host[:i]
			addr.BindAddr = host[i+1:]
		}
		if addr.BindAddr == "" || addr.AdvertiseAddr == "" {
			log.Panicf("invalid address: %s", addr)
		}
		result = append(result, addr)
	}
	return result
}

// MySQL配置
type MySQLConf struct {
	Addr     string
	User     string
	Password string
	Database string
}

func (c *MySQLConf) DSN() string {
	var cfg = mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = c.Addr
	cfg.User = c.User
	cfg.Passwd = c.Password
	cfg.DBName = c.Database
	cfg.ParseTime = true
	cfg.InterpolateParams = true
	cfg.Params["charset"] = "utf8mb4"
	return cfg.FormatDSN()
}
