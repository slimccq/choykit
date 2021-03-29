// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"encoding/json"
	"fmt"
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

//
const (
	RUNTIME_EXECUTOR_CAPACITY      = "RUNTIME_EXECUTOR_CAPACITY"
	RUNTIME_EXECUTOR_CONCURRENCY   = "RUNTIME_EXECUTOR_CONCURRENCY"
	RUNTIME_CONTEXT_INBOUND_SIZE   = "RUNTIME_CONTEXT_INBOUND_SIZE"
	RUNTIME_CONTEXT_OUTBOUND_SIZE  = "RUNTIME_CONTEXT_OUTBOUND_SIZE"
	RUNTIME_ENDPOINT_OUTBOUND_SIZE = "RUNTIME_ENDPOINT_OUTBOUND_SIZE"
	NET_PEER_PING_INTERVAL         = "NET_PEER_PING_INTERVAL"
	NET_PEER_READ_INTERVAL         = "NET_PEER_READ_INTERVAL"
	NET_RPC_TTL                    = "NET_RPC_TTL"
	NET_SESSION_READ_TIMEOUT       = "NET_SESSION_READ_TIMEOUT"
	NET_INTERFACES                 = "NET_INTERFACES"
)

// 进程的环境， 代码内部都使用environ获取变量参数
type Environ struct {
	protocol.Environ
	dotenv.Env
}

func (e *Environ) IsProd() bool {
	return e.AppEnv == "prod"
}

// command line option只是一种设置environ的手段
func (e *Environ) SetByOption(opt *Options) {
	if opt.WorkingDir != "" {
		e.AppWorkingDir = opt.WorkingDir
	}
	if opt.LogLevel != "" {
		e.AppLogLevel = opt.LogLevel
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
}

func NewEnviron() *Environ {
	return &Environ{
		Environ: protocol.Environ{
			AppEnv:       "dev",
			AppLogLevel:  "debug",
			EtcdAddr:     "127.0.0.1:2379",
			EtcdKeyspace: "choyd",
			EtcdLeaseTtl: 5,
		},
		Env: make(dotenv.Env),
	}
}

// 加载环境变量
func LoadEnviron() *Environ {
	env := NewEnviron()
	rv := reflect.ValueOf(env).Elem()
	rType := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		var name = strutil.ToSnakeCase(rType.Field(i).Name) // 'APP_ID' ==> 'AppId'
		var envKey = strings.ToUpper(name)
		var value = strings.TrimSpace(dotenv.Get(envKey))
		if value == "" {
			continue // 值为空的时候不覆盖
		}
		strutil.ParseStringToValue(value, rv.Field(i))
	}
	// 加载网络接口地址
	for _, iftext := range strings.Split(dotenv.Get(NET_INTERFACES), ",") {
		eth := ParseNetInterface(iftext)
		env.NetInterfaces = append(env.NetInterfaces, eth)
	}
	return env
}

func (e Environ) String() string {
	data, _ := json.Marshal(&e)
	return string(data)
}

type NetInterface protocol.InterfaceAddr

func (i NetInterface) Interface() string {
	return fmt.Sprintf("%s:%d", i.BindAddr, i.Port)
}

func (i NetInterface) AdvertiseInterface() string {
	return fmt.Sprintf("%s:%d", i.AdvertiseAddr, i.Port)
}

// 解析地址格式，对外地址@bind地址:端口，如example.com@0.0.0.0:9527
func ParseNetInterface(text string) *protocol.InterfaceAddr {
	addr := &protocol.InterfaceAddr{}
	i := strings.Index(text, "@")
	if i > 0 {
		addr.AdvertiseAddr = text[:i]
		text = text[i+1:]
	}
	host, port, err := net.SplitHostPort(text)
	if err != nil {
		log.Panicf("parse address %s: %v", text, err)
	}
	n, _ := strconv.Atoi(port)
	addr.Port = int32(n)
	addr.BindAddr = host
	if addr.BindAddr == "" {
		log.Panicf("invalid address: %s", addr)
	}
	if addr.AdvertiseAddr == "" {
		addr.AdvertiseAddr = addr.BindAddr
	}
	return addr
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
