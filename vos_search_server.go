package main

/*
	编译为linux环境：
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build vos_search_server.go
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/wxnacy/wgo/arrays"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	LOGPATH  = "/var/log/black/"
	FORMAT   = "20060102"
	LineFeed = "\r\n"
)

var path = LOGPATH + time.Now().Format(FORMAT) + "/"

//WriteLog return error
func WriteLog(fileName, msg string) error {
	if !IsExist(path) {
		return CreateDir(path)
	}
	var (
		err error
		f   *os.File
	)

	f, err = os.OpenFile(path+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	_, err = io.WriteString(f, LineFeed+msg)

	defer f.Close()
	return err
}

//CreateDir  文件夹创建
func CreateDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	os.Chmod(path, os.ModePerm)
	return nil
}

//IsExist  判断文件夹/文件是否存在  存在返回 true
func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

var (
	redis_client *redis.Client
)

func Redis_connect_server(ipaddr string) *redis.Client {
	var opt redis.Options
	opt.Addr = ipaddr
	opt.DB = 0
	redis_client = redis.NewClient(&opt)
	return redis_client
}

//VOS结构体
type einfo struct {
	CallId     int64  `json:"callId"`
	CallerE164 string `json:"callerE164"`
	CalleeE164 string `json:"calleeE164"`
}

type e164reg struct {
	RewriteE164Req einfo `json:"RewriteE164Req"`
}

type updatenum struct {
	MobileNum string `json:"mobilenum"`
	Nclass    int    `json:"nclass"`
}

//http服务函数
func init_http_client(port int) {
	http_server := http.NewServeMux()
	http_server.HandleFunc("/update", http_update_number) //上报号码接口
	http_server.HandleFunc("/bcheck", http_black_check)   //vos 重定项接口
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), http_server)
}

//黑名单服务函数
func http_black_check(w http.ResponseWriter, r *http.Request) {
	var numbers []string
	numbers = append(numbers, "13", "14", "15", "16", "17", "18", "19")
	ipaddr := strings.Split(r.RemoteAddr, ":")[0]

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("{\"code\":%d,\"message\":\"%s\"}", 0, "网络错误")))
		return
	}
	var res e164reg
	buf = bytes.TrimPrefix(buf, []byte("\xef\xbb\xbf"))
	err = json.Unmarshal(buf, &res)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("{\"code\":%d,\"message\":\"%s\"}", 0, "json解析数据不正确")))
		return
	}

	//fmt.Println("Query string: %s", string(buf))
	callId := res.RewriteE164Req.CallId
	callere164 := res.RewriteE164Req.CallerE164
	called := res.RewriteE164Req.CalleeE164

	//获取完请求数据
	fmt.Println("ipaddr", ipaddr, " - CallId:", callId, " CallerE164:", callere164, " CalleeE164:", called)

	//校验是否为手机号码
	if len(called) > 11 {
		called = called[len(called)-11:]
	}
	index := arrays.ContainsString(numbers, called[0:2]) // 前2位匹配数组
	if index == -1 {
		w.Write([]byte(fmt.Sprintf("{\"RewriteE164Rsp\":{\"callId\":%d,\"callerE164\":\"%s\",\"calleeE164\":\"ERRNUM%s\"}}", callId, callere164, res.RewriteE164Req.CalleeE164)))
		//fmt.Println("Called-非手机号")
		return
	}
	err = redis_client.Get(called).Err()
	if err != nil {
		w.Write([]byte(fmt.Sprintf("{\"RewriteE164Rsp\":{\"callId\":%d,\"callerE164\":\"%s\",\"calleeE164\":\"%s\"}}", callId, callere164, res.RewriteE164Req.CalleeE164)))
	} else {
		w.Write([]byte(fmt.Sprintf("{\"RewriteE164Rsp\":{\"callId\":%d,\"callerE164\":\"%s\",\"calleeE164\":\"Black_%s\"}}", callId, callere164, res.RewriteE164Req.CalleeE164)))
	}

	return
}

func http_update_number(w http.ResponseWriter, r *http.Request) {
	ipaddr := strings.Split(r.RemoteAddr, ":")[0]
	fmt.Println("ipaddress:", ipaddr)

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Read Body Err:", err)
	}
	var res updatenum
	buf = bytes.TrimPrefix(buf, []byte("\xef\xbb\xbf"))
	err = json.Unmarshal(buf, &res)
	if err != nil {
		fmt.Println("updatenum unmarshal json Err:", err)
		w.Write([]byte(fmt.Sprintf("{\"code\":%d,\"message\":\"%s\"}", 0, "参数不正确")))
		return
	}

	fmt.Println("Query string: ", string(buf))

	mobilenum := res.MobileNum
	nclass := res.Nclass

	if len(mobilenum) > 11 {
		mobilenum = mobilenum[len(mobilenum)-11:]
	}

	err = redis_client.Set(mobilenum, nclass, 0).Err()
	if err != nil {
		panic(err)
		w.Write([]byte(fmt.Sprintf("{\"code\":%d,\"message\":\"%s\"}", 0, "Write Redis Err")))
	} else {
		w.Write([]byte(fmt.Sprintf("{\"code\":%d,\"message\":\"%s\"}", 1, "write none")))
	}

	return
}

func main() {
	fmt.Println("starting......")
	redis_client = Redis_connect_server("127.0.0.1:6379")
	init_http_client(9200) //启动http生产
	//http_axb_request()				//手动测试接口
}
