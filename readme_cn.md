# go-http-agent

`go-http-agent` 是一个使用 Go 语言编写的 HTTP 代理程序。它支持反向代理功能，可将客户端的 HTTP 请求转发到指定的服务器，
并将服务器响应返回给客户端,并支持保存静态文件到本地。此代理程序支持 HTTP 和 HTTPS 协议，适用于多种网络环境和需求。

## 特性

- **反向代理**：可将客户端请求转发到任何 HTTP/HTTPS 服务器。
- **灵活配置**：通过命令行参数自定义端口、上下文路径和代理目标 URL。
- **HTTPS 支持**：能够处理 HTTPS 请求，适用于安全敏感的应用场景。
- **静态文件保存**：支持将访问的静态文件保存到本地指定目录。

## 启动命令

基本命令格式如下：

```shell
go-http-agent --port <端口号> --context-path <上下文路径> --proxy-url <目标代理URL>
```

### 示例

1. **反向代理百度**：

   ```shell
   go-http-agent --port 8080 --context-path / --proxy-url http://www.baidu.com
   ```

   成功设置后，反向代理百度如下图所示：
   ![反向代理百度示例](readme_files/1.jpg)

2. **自定义工程名**：

   ```shell
   go-http-agent --port 8080 --context-path /xxx --proxy-url http://www.baidu.com/yyy
   ```

   使用此配置，您可以通过自定义的上下文路径访问代理目标。

3. **HTTPS 代理**：

   ```shell
   go-http-agent --port 8080 --context-path / --proxy-url https://www.google.com
   ```

   此配置允许您代理 HTTPS 请求，例如代理到 Google。

## 参数说明

- `--port`：设置代理服务器的监听端口。
- `--context-path`：定义代理服务器的上下文路径。
- `--proxy-url`：指定需要代理到的目标 URL。
- `--save-dir`：指定静态文件的保存路径（用于保存访问过程中的静态资源）。
