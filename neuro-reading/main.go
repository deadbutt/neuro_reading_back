package main

import (
	"fmt"
	"neuro-reading/config"
	"neuro-reading/db"
	"neuro-reading/router"
)

func main() {
	cfg := config.Load()

	if err := db.Init(cfg); err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}

	r := router.Setup(cfg)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("服务器启动成功，监听地址: http://%s\n", addr)
	if err := r.Run(addr); err != nil {
		panic(fmt.Sprintf("服务器启动失败: %v", err))
	}
}