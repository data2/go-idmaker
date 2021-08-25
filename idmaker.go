package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	Ip         string
	CallTime   string
	ReturnTime string
}

type SeqId struct {
	id int32
	mu sync.RWMutex
}

type IdMaker struct {
	SeqId SeqId
}

type Response struct {
	id      int32
	code    int
	message string
}

func PrettyClient(c Client) {
	fmt.Println(fmt.Sprintf("请求对象信息,ip:%s, 调用时间:%s", c.Ip, c.CallTime))
}

func PrettyLockClient(c Client) {
	fmt.Println(fmt.Sprintf("请求对象信息,ip:%s, 调用时间:%s, 已经获取到锁正在处理", c.Ip, c.CallTime))
}

func PrettyClientReturn(c Client, id int32) {
	fmt.Println(fmt.Sprintf("请求对象信息,ip:%s, 调用时间:%s, 已经处理完毕释放锁，获取到值时间:%s, 返回序列id:%d", c.Ip, c.CallTime, c.ReturnTime, id))
}

func (r *Response) Str() string {
	return fmt.Sprintf("{'id':%d,'code':%d,'message':%s}", r.id, r.code, r.message)
}

func (im *IdMaker) GetSeqId() *SeqId {
	im.SeqId.mu.RLocker()
	seq := im.SeqId
	im.SeqId.mu.RUnlock()
	return &seq
}

func (im *IdMaker) GetNewSeqId(c Client) *SeqId {
	//PrettyClient(c)
	im.SeqId.mu.Lock()
	//PrettyLockClient(c)
	im.SeqId.id += 1
	seq := im.SeqId
	im.SeqId.mu.Unlock()
	PrettyClientReturn(c, seq.id)
	return &seq
}

var idMaker = &IdMaker{SeqId: SeqId{id: 0}}

func GetIp(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if len(ip) == 0 {
		ip = r.Header.Get("X-Forwarded-For")
	}
	return ip
}
func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, (&Response{id: idMaker.GetNewSeqId(Client{
		Ip: GetIp(r), CallTime: time.Now().String(),
	}).id, code: 200, message: "success"}).Str())

}
func main() {
	http.HandleFunc("/idMaker", index)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("启动失败")
	}

}
