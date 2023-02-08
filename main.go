/*
编写一个简单的 http代理服务
*/
package main

import (
  "log"
  "net/http"
  "net/http/httputil"
  "net/url"
  "os"
)

func main() {
  port := "3000"
  contextPath := "/"
  proxyUrl := ""
  for i := 1; i < len(os.Args); i += 2 {
    param := os.Args[i]
    if param == "--port" {
      port = os.Args[i+1]
    }
    if param == "--context-path" {
      contextPath = os.Args[i+1]
    }

    if param == "--proxy-url" {
      proxyUrl = os.Args[i+1]
    }
  }
  if proxyUrl == "" {
    log.Fatalln("please use --proxy-url to set proxy-url")
  }
  log.Println("start", port, contextPath, proxyUrl)
  remote, err := url.Parse(proxyUrl) //远程服务器
  if err != nil {
    panic(err)
  }
  //
  proxy := httputil.NewSingleHostReverseProxy(remote)

  http.HandleFunc(contextPath, handler(proxy)) //注册路由
  log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    log.Println(r.URL)
    r.Host = r.URL.Host
    p.ServeHTTP(w, r)
  }
}
