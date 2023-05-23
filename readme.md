# go-http-agent
使用go语言编写的http代理程序

启动命令
```shell
go-http-agent --port 8080 --context-path / --proxy-url http://www.baidu.com
```


成功反向代理百度
![](readme_files/1.jpg)

支持自定义工程名

```
go-http-agent --port 8080 --context-path /xxx --proxy-url http://www.baidu.com/yyy
```
支持https代理
```
go-http-agent --port 8080 --context-path / --proxy-url https://www.google.com
```
