/*
编写一个简单的 http代理服务
*/
package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// 全局变量来存储静态文件扩展名
var staticFileExtensions map[string]bool

func main() {
	// 解析命令行参数
	port := flag.String("port", "3000", "Port to run the server on")
	contextPath := flag.String("context-path", "/", "Context path of the proxy server")
	proxyURL := flag.String("proxy-url", "", "URL of the proxy server")
	saveDir := flag.String("save-dir", "./saved_files", "Directory to save static files")
	flag.Parse()

	// 读取静态文件扩展名
	staticFileExtensions = make(map[string]bool)
	readStaticFileExtensions("static_file.txt")
	remote, err := url.Parse(*proxyURL) //远程服务器
	if err != nil {
		panic(err)
	}
	//
	proxy := httputil.NewSingleHostReverseProxy(remote)
	http.HandleFunc(*contextPath, handler(proxy, *saveDir, remote))
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func handler(p *httputil.ReverseProxy, saveDir string, proxyURL *url.URL) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// 检查请求的URL是否指向静态文件
		if saveDir != "" && isStaticFile(r.URL.Path) {
			log.Println("save ", r.URL)
			saveStaticFile(r, saveDir, proxyURL)
		} else {
			log.Println(r.URL)
		}

		r.Host = r.URL.Host
		p.ServeHTTP(w, r)
	}
}

// 读取静态文件扩展名
func readStaticFileExtensions(filePath string) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件不存在，创建并写入扩展名
		createAndWriteStaticExtensions(filePath)
	}

	// 读取文件中的扩展名
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening static_file.txt: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		staticFileExtensions[scanner.Text()] = true
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading static_file.txt: %v", err)
	}
}

func createAndWriteStaticExtensions(filePath string) {
	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Error creating static_file.txt: %v", err)
	}
	defer file.Close()

	// 已知的静态文件扩展名
	extensions := []string{".css", ".js", ".jpg", ".jpeg", ".png", ".svg", ".webp", ".html", ".htm", ".txt", ".md", ".pdf"}

	// 写入扩展名
	for _, ext := range extensions {
		if _, err := file.WriteString(ext + "\n"); err != nil {
			log.Fatalf("Error writing to static_file.txt: %v", err)
		}
	}
}

func isStaticFile(path string) bool {
	for ext := range staticFileExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

// 保存静态文件
func saveStaticFile(r *http.Request, saveDir string, proxyURL *url.URL) {
	// 构造完整的远程URL
	remoteURL := proxyURL.ResolveReference(r.URL).String()

	// 构造文件保存路径
	filePath := saveDir + r.URL.Path

	// 确保保存文件的目录存在
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		log.Println("Error creating directory:", err)
		return
	}

	// 获取文件内容
	resp, err := http.Get(remoteURL)
	if err != nil {
		log.Println("Error fetching static file:", err)
		return
	}
	defer resp.Body.Close()

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		log.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// 写入文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Println("Error saving file:", err)
		return
	}
}
