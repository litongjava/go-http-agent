package main

import (
  "bufio"
  "compress/flate"
  "compress/gzip"
  "flag"
  "github.com/andybalholm/brotli"
  "github.com/cloudwego/hertz/pkg/common/hlog"
  "github.com/gorilla/websocket"
  "io"
  "io/ioutil"
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

  hlog.SetLevel(hlog.LevelInfo)
  f, err := os.Create("app.log")
  if err != nil {
    panic(err)
  }
  defer f.Close()

  fileWriter := io.MultiWriter(f, os.Stdout)
  hlog.SetOutput(fileWriter)

  // 读取静态文件扩展名
  staticFileExtensions = make(map[string]bool)
  readStaticFileExtensions("static_file.txt")
  remote, err := url.Parse(*proxyURL) //远程服务器
  if err != nil {
    panic(err)
  }
  //
  proxy := httputil.NewSingleHostReverseProxy(remote)
  http.HandleFunc(*contextPath, ProxyHandler(proxy, *saveDir, remote))
  log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func ProxyHandler(reverseProxy *httputil.ReverseProxy, saveDir string, proxyURL *url.URL) func(http.ResponseWriter, *http.Request) {
  return func(response http.ResponseWriter, request *http.Request) {

    // 检查是否是 WebSocket 请求
    if websocket.IsWebSocketUpgrade(request) {
      handleWebSocket(proxyURL, response, request)
      return
    }

    // 检查请求的URL是否指向静态文件
    if saveDir != "" && isStaticFile(request.URL.Path) {
      hlog.Info("save ", request.URL)
      saveStaticFile(request, saveDir, proxyURL)
    } else {
      hlog.Info(request.URL)
    }

    request.Host = request.URL.Host
    // 捕获和打印响应头和响应体
    captureAndLogResponse(response, request, reverseProxy)
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

var upgrader = websocket.Upgrader{
  ReadBufferSize:  1024,
  WriteBufferSize: 1024,
}

func handleWebSocket(proxyURL *url.URL, w http.ResponseWriter, r *http.Request) {
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Println("Error upgrading to websocket:", err)
    return
  }
  defer conn.Close()

  // 连接到远程 WebSocket 服务器
  remoteConn, _, err := websocket.DefaultDialer.Dial(proxyURL.String(), nil)
  if err != nil {
    log.Println("Error dialing remote websocket:", err)
    return
  }
  defer remoteConn.Close()

  // 处理从客户端到服务器的消息
  go func() {
    for {
      messageType, message, err := conn.ReadMessage()
      if err != nil {
        log.Println("Error reading from websocket:", err)
        break
      }
      err = remoteConn.WriteMessage(messageType, message)
      if err != nil {
        log.Println("Error writing to remote websocket:", err)
        break
      }
    }
  }()

  // 处理从服务器到客户端的消息
  for {
    messageType, message, err := remoteConn.ReadMessage()
    if err != nil {
      log.Println("Error reading from remote websocket:", err)
      break
    }
    err = conn.WriteMessage(messageType, message)
    if err != nil {
      log.Println("Error writing to websocket:", err)
      break
    }
  }
}

func logRequest(r *http.Request) {
  hlog.Info("Request URL:", r.URL)
  hlog.Info("Request Headers:", r.Header)
  if r.Method == http.MethodPost || r.Method == http.MethodPut {
    bodyBytes, err := io.ReadAll(r.Body)
    if err != nil {
      hlog.Error("Error reading request body:", err)
    } else {
      if len(bodyBytes) > 1024*100 {
        hlog.Info("Request Body more that 100k")
      } else {
        hlog.Error("Request Body:", string(bodyBytes))
      }
      r.Body = io.NopCloser(strings.NewReader(string(bodyBytes))) // 重新构建请求体
    }
  }
}

func captureAndLogResponse(response http.ResponseWriter, request *http.Request, reverseProxy *httputil.ReverseProxy) {
  // 打印请求头和请求体
  logRequest(request)
  rw := &responseWriter{ResponseWriter: response}
  reverseProxy.ServeHTTP(rw, request)
  hlog.Info("Response Headers:", rw.Header())
  decodedBody, err := decodeResponseBody(rw.Header(), rw.body)
  if err != nil {
    hlog.Error("Error decoding response body:", err)
    return
  }
  length := len(decodedBody)
  if length > 1024*100 {
    hlog.Info("Response Body more than 100k")
  } else {
    hlog.Info("Response Body:", string(decodedBody))
  }
}

func decodeResponseBody(headers http.Header, body []byte) ([]byte, error) {
  contentEncoding := headers.Get("Content-Encoding")
  var reader io.Reader
  reader = strings.NewReader(string(body))

  switch contentEncoding {
  case "gzip":
    gz, err := gzip.NewReader(reader)
    if err != nil {
      return nil, err
    }
    defer gz.Close()
    return ioutil.ReadAll(gz)
  case "deflate":
    fl := flate.NewReader(reader)
    defer fl.Close()
    return ioutil.ReadAll(fl)
  case "br":
    br := brotli.NewReader(reader)
    return ioutil.ReadAll(br)
  default:
    return body, nil
  }
}

type responseWriter struct {
  http.ResponseWriter
  body []byte
}

func (rw *responseWriter) Write(b []byte) (int, error) {
  rw.body = append(rw.body, b...)
  return rw.ResponseWriter.Write(b)
}
