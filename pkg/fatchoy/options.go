// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"devpkg.work/choykit/pkg/x/dotenv"
	"devpkg.work/choykit/pkg/x/strutil"
	"github.com/jessevdk/go-flags"
)

// 命令行选项参数
type Options struct {
	ShowVersion       bool   `short:"v" long:"version" description:"version string"`
	List              bool   `short:"l" long:"list" description:"list available services"`
	ConfigFile        string `short:"C" long:"config" description:"load option from file"`
	EnvFile           string `short:"E" long:"envfile" description:"dotenv file path"`
	WorkingDir        string `short:"W" long:"workdir" description:"runtime working directory"`
	ResourceDir       string `short:"R" long:"resdir" description:"resource directory"`
	ServiceType       string `short:"S" long:"service-type" description:"name of service type"`
	ServiceIndex      int16  `short:"N" long:"service-index" description:"instance index of this service"`
	ServiceDependency string `short:"P" long:"dependency" description:"service dependency list"`
	Interface         string `short:"I" long:"interface" description:"service interface address"`
	LogLevel          string `short:"L" long:"loglevel" description:"debug,info,warn,error,fatal,panic"`
	EtcdAddress       string `short:"F" long:"etcd-addr" description:"etcd host address"`
	EtcdKeySpace      string `short:"K" long:"keyspace" description:"etcd key prefix"`
	EtcdLeaseTTL      int    `long:"lease-ttl" description:"etcd lease key TTL"`
	PprofAddr         string `long:"pprof-addr" description:"pprof http listen address"`
	EnableSysLog      bool   `long:"enable-syslog" description:"enable write log to syslog/eventlog"`
	SysLogParams      string `long:"syslog-param" description:"syslog/eventlog parameters"`
}

func NewOptions() *Options {
	return &Options{
		LogLevel:     "debug",
		ResourceDir:  "res",
		EnvFile:      "etc/.env",
		EtcdKeySpace: "choyd",
		EtcdAddress:  "127.0.0.1:2379", // local etcd instance
	}
}

// Parse options from console
func ParseOptions() (*Options, error) {
	var opts = NewOptions()
	if _, err := flags.Parse(opts); err != nil {
		if e, ok := err.(*flags.Error); ok {
			if e.Type == flags.ErrHelp {
				return nil, nil
			}
		}
		return nil, nil
	}
	if opts.WorkingDir == "" {
		cwd, _ := os.Getwd()
		opts.WorkingDir = cwd
	}
	opts.ServiceType = strings.ToLower(opts.ServiceType)
	return opts, nil
}

// 从文件中读取option
func ReadFileOption(filename string, opts *Options) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	env, err := dotenv.ParseEnv(string(data))
	if err != nil {
		return err
	}
	strutil.ParseKVToStruct(env, reflect.ValueOf(opts).Elem())

	return nil
}
