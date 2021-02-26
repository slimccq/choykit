// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"encoding/json"

	"devpkg.work/choykit/pkg/x/dotenv"
	"github.com/go-sql-driver/mysql"
)

// 基本概念
//   游戏: 最顶级的单位
//   大世界: 一个游戏下可以有一个或多个大世界，比如常见的分全球版大世界和TW-BETA版大世界
//   游戏版本: 一个大世界下可以有一个或多个游戏版本，游戏版本的主要构成以平台+地区标识为粒度，比如IOS-EN / AMZ-FR
//   游戏服务器: 用于大世界下面的一种逻辑分服

// 通用环境变量
type Environ struct {
	DevelopMode               bool   // 测试/生产环境
	GameID                    string // 游戏ID
	ChannelID                 string // 渠道ID
	ServerID                  string // 服务器ID
	ServerName                string // 服务器名称
	AccessKey                 string //
	MysqlDSN                  string //
	RedisAddr                 string //
	ExecutorCapacity          int    //
	ExecutorConcurrency       int    //
	ContextInboundQueueSize   int    //
	ContextOutboundQueueSize  int    //
	EndpointOutboundQueueSize int    //
	NetEnableEncryption       bool
	NetPublicKeyFile          string //
	NetPrivateKeyFile         string //
	NetPeerPingInterval       int    //
	NetPeerReadTimeout        int    //
	NetSessionReadTimeout     int    //
	NetRpcTimeoutInterval     int    //
}

func NewEnviron() *Environ {
	return &Environ{
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

func LoadEnviron() *Environ {
	env := NewEnviron()
	env.DevelopMode = dotenv.GetBool("APP_DEVELOP_MODE")
	env.GameID = dotenv.Get("APP_GAME_ID")
	env.ChannelID = dotenv.Get("APP_CHANNEL_ID")
	env.ServerID = dotenv.Get("APP_SERVER_NAME")
	env.ServerID = dotenv.Get("APP_SERVER_ID")
	env.AccessKey = dotenv.Get("APP_ACCESS_KEY")
	env.MysqlDSN = dotenv.Get("DB_MYSQL_DSN")
	env.RedisAddr = dotenv.Get("DB_REDIS_ADDR")
	env.ExecutorCapacity = dotenv.GetInt("RUNTIME_EXECUTOR_CAPACITY")
	env.ContextInboundQueueSize = dotenv.GetInt("RUNTIME_CONTEXT_INBOUND_SIZE")
	env.ContextOutboundQueueSize = dotenv.GetInt("RUNTIME_CONTEXT_OUTBOUND_SIZE")
	env.EndpointOutboundQueueSize = dotenv.GetInt("RUNTIME_ENDPOINT_OUTBOUND_SIZE")
	env.NetEnableEncryption = dotenv.GetBool("NET_ENABLE_ENCRYPTION")
	env.NetPublicKeyFile = dotenv.Get("NET_PUBKEY_FILE")
	env.NetPrivateKeyFile = dotenv.Get("NET_PRIKEY_FILE")
	env.NetRpcTimeoutInterval = dotenv.GetInt("NET_RPC_TIMEOUT_INTERVAL")
	env.NetPeerPingInterval = dotenv.GetInt("NET_PEER_PING_INTERVAL")
	env.NetPeerReadTimeout = dotenv.GetInt("NET_PEER_READ_TIMEOUT")
	env.NetSessionReadTimeout = dotenv.GetInt("NET_SESSION_READ_TIMEOUT")
	return env
}

func (e Environ) String() string {
	data, _ := json.Marshal(&e)
	return string(data)
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
