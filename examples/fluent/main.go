// Copyright © 2020 Hedzr Yeh.

package main

import (
	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	svcConfig := &service.Config{
		Name:        "hz-fluent",              // 服务显示名称
		DisplayName: "HZ-Fluent",              // 服务名称
		Description: "hz-fluent - nt service", // 服务描述
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}

	if err != nil {
		logrus.Fatal(err)
	}

	if len(os.Args) > 1 {
		if os.Args[1] == "install" {
			s.Install()
			logrus.Println("服务安装成功")
			return
		}

		if os.Args[1] == "remove" {
			s.Uninstall()
			logrus.Println("服务卸载成功")
			return
		}
	}

	err = s.Run()
	if err != nil {
		logrus.Error(err)
	}
}

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	// 代码写在这儿
}

func (p *program) Stop(s service.Service) error {
	return nil
}
