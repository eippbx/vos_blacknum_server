## vos_black_search
vos黑名单查询缓存
> 运行环境为Centos 6x/7x Redis

##VOS3000客户机：
```不需要安装环境。```
### vos机器运行的程序编译：
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build vos_search_up.go
```

> 功能： 查询vos3000的cdr文件内被叫有Black标志的数据上报到服务端，加入redis库里备查

> 运行：手动指定日期  ./vos_search_up -date 20220502

#### 将vos_search_up 和 vos_search_up.sh 放入/root目录，并给777执行权限

#### 加到crontab服务里，设定每天早上1点进行上报
```bash
0 1 * * * /root/vos_search_up.sh
```

##黑名单服务器：
```yum -y install redis```

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
