package main

//	连接VOS没问题
//	编译 :
//	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build vos_search_up.go
import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Cdr struct {
	Id         uint64
	Calleee164 string
	Num        string
}

var date = flag.String("date", "20220501", "日期格式:20220501")

/*
	linux shell获取前一天的日期：date -d yesterday +%Y%m%d
	shell:
		ydate=$(date -d "yesterday" +%Y%m%d)
		echo ${ydate}
		./vos_seach -date ${ydate}
*/

func main() {

	var cdrs []Cdr
	var u_class int = 0
	db, err := gorm.Open("mysql", "root:@(127.0.0.1:3306)/vos3000?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
		fmt.Println("连接数据库失败")
		return
	}
	flag.Parse()
	fmt.Println("date ", *date)

	//sql := "select id,calleee164,calleeaccesse164 as num from e_cdr_"+yestoday+" where calleee164 like \"BlackNum%\" limit 10 "
	sql := "select id,calleee164,calleeaccesse164 as num from e_cdr_" + *date + " where calleee164 like \"BlackNum%\" "
	db.Raw(sql).Scan(&cdrs)
	fmt.Println("SQL:", sql)
	for _, v := range cdrs {
		num_class := v.Calleee164[8:9]
		switch {
		case num_class == "G":
			u_class = 3
		case num_class == "Z":
			u_class = 2
		case num_class == "D":
			u_class = 1
		default:
			u_class = 0
		}
		fmt.Println("ID:", v.Id, "   Calleee164:", v.Calleee164, "   num:", v.Num, "    num_class:", num_class, "  u_class:", strconv.Itoa(u_class))
		if len(v.Num) > 11 {
			v.Num = v.Num[len(v.Num)-11:]
		}
		code, message := http_up_number(v.Num, u_class)
		fmt.Println("re Code:", code, " message:", message)
	}

	defer db.Close()
}

func http_up_number(mobilenum string, nclass int) (int, string) {
	type Resultmsg_str struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	type Post_str struct {
		MobileNum string `json:"mobilenum"`
		Nclass    int    `json:"nclass"`
	}
	post_str := Post_str{
		MobileNum: mobilenum,
		Nclass:    nclass,
	}

	b, err := json.Marshal(post_str)
	if err != nil {
		return 0, "line 71"
	}
	body := bytes.NewBuffer(b)
	//fmt.Println(body)

	addr := "http://39.103.219.47:9200/update"
	contentType := "application/json;charset=utf-8"
	client := &http.Client{}
	req, err := http.NewRequest("POST", addr, body)
	if err != nil {
		return 0, "line 80"
	}

	req.Header.Set("Content-Type", contentType)
	resq, err := client.Do(req)
	if err != nil {
		return 0, "line 85"
	}
	buf, err := ioutil.ReadAll(resq.Body)
	resq.Body.Close()
	if err != nil {
		return 0, "line 90"
	}
	fmt.Println(string(buf))

	var res Resultmsg_str
	buf = bytes.TrimPrefix(buf, []byte("\xef\xbb\xbf"))
	err = json.Unmarshal(buf, &res)
	if err != nil {
		return 0, "line 98"
	}

	return res.Code, res.Message
}
