package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"web_music/api"
)

func main() {
	// 设置日志
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("云音乐服务启动中...")

	// 设置路由
	setupRoutes()

	// 获取端口，如果环境变量没有设置则使用默认端口8081
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	// 启动服务器
	log.Printf("服务器启动在 http://localhost:%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// 设置路由
func setupRoutes() {
	// 静态文件服务
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// API路由
	http.HandleFunc("/api/search", api.SearchHandler)
	http.HandleFunc("/api/song", api.SongHandler)

	// 主页
	http.HandleFunc("/", indexHandler)
}

// 主页处理函数，返回index.html
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 获取index.html文件的绝对路径
	indexPath := filepath.Join("templates", "index.html")

	// 读取文件内容
	content, err := os.ReadFile(indexPath)
	if err != nil {
		log.Printf("读取index.html失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}

	// 设置内容类型并发送响应
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}
