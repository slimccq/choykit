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
	fatchoy.StartClock()

	workDir := opts.WorkingDir
	if workDir != "" {
		if err := os.Chdir(workDir); err != nil {
			return err
		}
	} else {
		workDir, _ = os.Getwd()
	}

	// load .env variable
	if opts.EnvFile != "" && fsutil.IsFileExist(opts.EnvFile) {
		if err := dotenv.Load(opts.EnvFile, true); err != nil {
			return err
		}
	}
	env := fatchoy.LoadEnviron()
	env.SetByOption(opts)

	var srv = fatchoy.GetServiceByName(opts.ServiceType)
	if srv == nil {
		return errors.Errorf("unrecognized service [%s]", opts.ServiceType)
	}
	var node = fatchoy.MakeNodeID(srv.ID(), opts.ServiceIndex)
	srv.SetNodeID(node)

	os.Mkdir("logs", 0755)
	var filepath = fmt.Sprintf("logs/%s_%d.log", srv.Name(), opts.ServiceIndex)
	log.Setup(env.IsProd(), opts.EnableSysLog, opts.LogLevel, filepath, opts.SysLogParams)

	log.Infof("working dir: %s", workDir)
	log.Infof("service type: %s", env.ServiceType)
	log.Infof("service index: %d", env.ServiceIndex)
	log.Infof("service dependency: %s", env.ServiceDependency)

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
	var ctx = fatchoy.NewServiceContext(d.env)
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
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	for {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL:
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
