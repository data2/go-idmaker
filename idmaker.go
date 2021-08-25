package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
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
	PrettyClient(c)
	im.SeqId.mu.Lock()
	PrettyLockClient(c)
	im.SeqId.id += 1
	seq := im.SeqId
	im.SeqId.mu.Unlock()
	c.ReturnTime = time.Now().String()
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

func PathExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func Load() {
	if !PathExist("~/.idMarker") {
		return
	}
	bytes, err := ioutil.ReadFile("~/.idMarker")
	if err != nil {
		panic(err)
	}
	if len(string(bytes)) != 0 {
		contentId, err := strconv.ParseInt(string(bytes), 10, 32)
		if err != nil {
			panic(err)
		}
		idMaker.SeqId.id = int32(contentId)
	}
}

func (im *IdMaker) Record() {
	id := im.SeqId.id
	fmt.Println("------------------------------------------------------------------------------------")
	fmt.Println(fmt.Sprintf("program execution done, record id : %d, time : %s", id, time.Now().String()))
	fmt.Println("------------------------------------------------------------------------------------")
	ioutil.WriteFile("~/.idMarker", []byte(strconv.Itoa(int(id))), 0664)
}

func BeautyExit(ch chan os.Signal) {
	fmt.Println("------------------------------------------------------------------------------------")
	fmt.Println("program execution begin , listening health...")
	fmt.Println("------------------------------------------------------------------------------------")
	for s := range ch {
		fmt.Println("------------------------------------------------------------------------------------")
		fmt.Println(fmt.Sprintf("program execution exit, receive signal：%s", s))
		fmt.Println("------------------------------------------------------------------------------------")
		idMaker.Record()
	}
}

func main() {
	Load()
	http.HandleFunc("/idMaker", index)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go BeautyExit(ch)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("启动失败")
	}

}
