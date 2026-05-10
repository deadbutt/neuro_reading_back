package main

import (
	"fmt"
	"neuro-reading/config"
	"neuro-reading/db"
	"neuro-reading/router"
	"os"
)

func main() {
	cfg := config.Load()

	if cfg.JWT.Secret == "neuro-reading-secret-key-change-in-production" {
		fmt.Println("警告: JWT_SECRET 未配置，使用了默认值，请在生产环境中设置环境变量 JWT_SECRET")
		fmt.Println("提示: 可以在 .env 文件中添加 JWT_SECRET=your-secret-key")
	}

	if err := db.Init(cfg); err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}

	r := router.Setup(cfg)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("服务器启动成功，监听地址: http://%s\n", addr)
	if err := r.Run(addr); err != nil {
		fmt.Fprintf(os.Stderr, "服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}
