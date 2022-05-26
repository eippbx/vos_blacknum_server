## 共有两个应用程序：
1、vos_search_server   #提供地址接收第三方过滤后的黑名单号码入库，同时为vos软交换提供黑名单服务。
2、vos_search_up       #安装在vos3000服务器上，定时提取号码上报到vos_search_server服务端

> 运行环境为Centos 6x/7x

### vos3000机器上运行的客户端vos_search_up程序的编译：
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build vos_search_up.go
```
> 功能： 查询vos3000的cdr文件内被叫有Black标志的数据上报到服务端，加入redis库里备查

> 如何运行：
> 手动指定日期  ./vos_search_up -date 20220502
> 加到crontab服务里，设定每天早上1点进行上报

```bash
1 1 * * * /root/vos_search_up.sh
```

### 服务端程序编译：
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build vos_search_server.go
```
> 执行:  ./vos_search_server
> 功能： 起服务监听9200端口，设置vos的url攺写，指向该服务进行黑名单查询
```bash
copy vos_serach_server.init /etc/init.d/vos_serach_server
chkconfig --add vos_search_server
chkconfig vos_search_server on
systemctl enable vos_serach_server
```
