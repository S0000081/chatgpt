package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
)

const (
	token = "sixah" //设置token
)

func makeSignature(timestamp, nonce string) string { //本地计算signature
	si := []string{token, timestamp, nonce}
	sort.Strings(si)            //字典序排序
	str := strings.Join(si, "") //组合字符串
	s := sha1.New()             //返回一个新的使用SHA1校验的hash.Hash接口
	io.WriteString(s, str)      //WriteString函数将字符串数组str中的内容写入到s中
	return fmt.Sprintf("%x", s.Sum(nil))
}

func validateUrl(w http.ResponseWriter, r *http.Request) bool {
	timestamp := strings.Join(r.Form["timestamp"], "")
	nonce := strings.Join(r.Form["nonce"], "")
	signature := strings.Join(r.Form["signature"], "")
	echostr := strings.Join(r.Form["echostr"], "")
	signatureGen := makeSignature(timestamp, nonce)

	if signatureGen != signature {
		return false
	}
	fmt.Fprintf(w, echostr) //原样返回eechostr给微信服务器
	return true
}

func procSignature(w http.ResponseWriter, r *http.Request) {

	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	log.Println("Wechat Service: Receive Msg: ", string(b))

}

func main() {
	log.Println("Wechat Service: Start!")
	http.HandleFunc("/notifyMsg", procSignature)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Println("Wechat Service: ListenAndServe Error: ", err)
	}
	log.Println("Wechat Service: Stop!")
}

//func main() {
//	if err := db.Init(); err != nil {
//		panic(fmt.Sprintf("mysql init failed with %+v", err))
//	}
//	g := gin.Default()
//	g.GET("/", func(c *gin.Context) {
//
//		signature := c.Query("signature")
//		timestamp := c.Query("timestamp")
//		nonce := c.Query("nonce")
//		echostr := c.Query("echostr")
//		token := "yuanshimima"
//		if signature == "" || timestamp == "" || nonce == "" || echostr == "" {
//			c.String(200, "参数错误")
//			return
//		}
//
//	})
//
//	//http.HandleFunc("/", service.IndexHandler)
//	//http.HandleFunc("/api/count", service.CounterHandler)
//	//
//	//log.Fatal(http.ListenAndServe(":80", nil))
//}
