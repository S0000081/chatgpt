package main

import (
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
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

type Body struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   string `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Content      string `xml:"Content"`
	MsgId        string `xml:"MsgId"`
}

// Reply 消息回复结构体
type Reply struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   string   `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content"`
}

type Msg struct {
	Model            string  `json:"model"`
	Prompt           string  `json:"prompt"`
	Temperature      float64 `json:"temperature"`
	MaxTokens        int64   `json:"max_tokens"`
	TopP             float64 `json:"top_p"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
}

type Text struct {
	Text string `json:"text"`
}

type gptMsg struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Choices []Text `json:"choices"`
}

func procSignature(w http.ResponseWriter, r *http.Request) {

	b, err := io.ReadAll(r.Body)
	log.Println(err)
	defer r.Body.Close()
	var body Body
	err = xml.Unmarshal(b, &body)

	if body.MsgType != "text" {
		reply := Reply{
			ToUserName:   body.FromUserName,
			FromUserName: body.ToUserName,
			CreateTime:   strconv.FormatInt(time.Now().Unix(), 10),
			MsgType:      "text",
			Content:      "暂时只支持文本消息😯",
		}
		bReply, _ := xml.Marshal(reply)
		w.Header().Set("Content-Type", "application/xml")
		w.Write(bReply)
		return
	}

	msg := Msg{
		Model:            "text-davinci-003",
		Prompt:           body.Content,
		Temperature:      0.5,
		MaxTokens:        60,
		TopP:             1.0,
		FrequencyPenalty: 0.5,
		PresencePenalty:  0.0,
	}
	bMsg, _ := json.Marshal(msg)

	request, _ := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/completions", strings.NewReader(string(bMsg)))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+os.Getenv("OPEN_KEY"))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()
	b, err = io.ReadAll(response.Body)
	log.Println(string(b))
	gptMsg := gptMsg{}
	err = json.Unmarshal(b, &gptMsg)
	if err != nil || len(gptMsg.Choices) == 0 {
		log.Println(err)
		return
	}
	reply := Reply{
		ToUserName:   body.FromUserName,
		FromUserName: body.ToUserName,
		CreateTime:   strconv.FormatInt(time.Now().Unix(), 10),
		MsgType:      "text",
		Content:      strings.ReplaceAll(gptMsg.Choices[0].Text, "\n", ""),
	}
	bReply, _ := xml.Marshal(reply)
	w.Header().Set("Content-Type", "application/xml")
	w.Write(bReply)
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
