// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"devpkg.work/choykit/pkg/bootstrap"
	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
)

func init() {
	rand.Seed(int64(os.Getpid()) ^ time.Now().UnixNano())
	http.DefaultClient.Timeout = time.Minute
}

func main() {
	var opts = parseOptions()
	var program bootstrap.Program
	if err := program.Init(opts); err != nil {
		log.ServerErrorLog("init bootstrap: %v", err)
		os.Exit(1)
	}
	if err := program.Run(); err != nil {
		log.ServerErrorLog("run service: %v", err)
		os.Exit(1)
	}
}

// 解析出一个option
func parseOptions() *fatchoy.Options {
	opts, err := fatchoy.ParseOptions()
	if err != nil {
		log.ServerErrorLog("ParseOptions: %v", err)
		os.Exit(1)
	}
	if opts == nil {
		os.Exit(0)
	}
	if opts.ShowVersion {
		fmt.Printf("version 0.0.1\n")
		os.Exit(0)
	}
	if opts.List {
		fmt.Printf("available service names:\n")
		for _, name := range fatchoy.GetServiceNames() {
			fmt.Printf("\t%s\n", name)
		}
		os.Exit(0)
	}
	if opts.ConfigFile != "" {
		if err := fatchoy.ReadFileOption(opts.ConfigFile, opts); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
			return nil
		}
	}
	return opts
}
