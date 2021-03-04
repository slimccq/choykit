// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package bootstrap

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"devpkg.work/choykit/pkg/fatchoy"
	"devpkg.work/choykit/pkg/log"
	"devpkg.work/choykit/pkg/x/dotenv"
	"devpkg.work/choykit/pkg/x/fsutil"
	"github.com/kardianos/service"
	"github.com/pkg/errors"
)

type Program struct {
	daemon service.Service
	logger service.Logger
	opts   *fatchoy.Options
	env    *fatchoy.Environ
	ctx    *fatchoy.ServiceContext
	app    fatchoy.Service
}

func (d *Program) Init(opts *fatchoy.Options) error {
	// create necessary dirs
	if err := os.Chdir(opts.WorkingDir); err != nil {
		return err
	}
	var dirNames = []string{"data", "log"}
	for _, name := range dirNames {
		os.Mkdir(name, 0755)
	}

	fatchoy.StartClock()

	// load .env variable
	if opts.EnvFile != "" && fsutil.IsFileExist(opts.EnvFile) {
		if err := dotenv.Load(opts.EnvFile, true); err != nil {
			return err
		}
	}
	env := fatchoy.LoadEnviron()

	var srv = fatchoy.GetServiceByName(opts.ServiceType)
	if srv == nil {
		return errors.Errorf("unrecognized service [%s]", opts.ServiceType)
	}
	var node = fatchoy.MakeNodeID(srv.ID(), opts.ServiceIndex)
	srv.SetNodeID(node)

	var filepath = fmt.Sprintf("log/%s_%d.log", srv.Name(), opts.ServiceIndex)
	log.Setup(!env.DevelopMode, opts.EnableSysLog, opts.LogLevel, filepath, opts.SysLogParams)

	// 注册protobuf消息反射
	// protocol.InitMessageRegistry()

	var cfg = &service.Config{
		Name:             "Choyd",
		DisplayName:      "Service Launcher",
		Description:      "Launch game service with this launcher program",
		Arguments:        os.Args[1:],
		WorkingDirectory: opts.WorkingDir,
	}
	daemon, err := service.New(d, cfg)
	if err != nil {
		return err
	}
	logger, err := daemon.Logger(nil)
	if err != nil {
		log.ServerErrorLog("cannot create daemon logger: %v", err)
		return err
	}

	logger.Infof("%s service(pid %d) initialized", opts.ServiceType, os.Getpid())

	d.opts = opts
	d.env = env
	d.app = srv
	d.logger = logger
	d.daemon = daemon

	if opts.PprofAddr != "" {
		go d.profiler(opts.PprofAddr)
	}

	return nil
}

func (d *Program) Run() error {
	return d.daemon.Run()
}

func (d *Program) Start(srv service.Service) error {
	var ctx = fatchoy.NewServiceContext(d.opts, d.env)
	if err := ctx.Start(d.app); err != nil {
		return err
	}
	go d.signaler()

	d.logger.Infof("%s service started", d.app.Name())

	d.daemon = srv
	d.ctx = ctx
	return nil
}

func (d *Program) doCleanJob() {
	d.ctx.Shutdown()
	log.Shutdown()
	fatchoy.StopClock()
}

func (d *Program) Stop(srv service.Service) error {
	d.doCleanJob()
	d.logger.Infof("%s service stopped", d.app.Name())
	return nil
}

func (d *Program) signaler() {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL)
	for {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGINT, syscall.SIGKILL:
				d.logger.Infof("signal %s received, start shutdown %s service",
					sig, d.app.Name())
				d.doCleanJob()
			}
		}
	}
}

func (d *Program) profiler(addr string) {
	d.logger.Infof("pprof serving at http://%s/debug/pprof", addr)
	http.ListenAndServe(addr, nil)
}
