package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"runtime"
	"strconv"
	"strings"
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
	im.Record()
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
	if !PathExist(PathJoin()) {
		return
	}
	bytes, err := ioutil.ReadFile(PathJoin())
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

func homeUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

var idMakerFile = "idMaker.txt"

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}

func GetHomePath() (string, string) {
	user, err := user.Current()
	if nil == err {
		return user.HomeDir, "linux"
	}

	// cross compile support
	if "windows" == runtime.GOOS {
		pathWindows, _ := homeWindows()
		return pathWindows, "windows"
	}

	// Unix-like system, so just assume Unix
	pathUnix, _ := homeUnix()
	return pathUnix, "unix"
}

func PathJoin() string {
	dir, osType := GetHomePath()
	switch osType {
	case "windows":
		return dir + "\\" + idMakerFile
	default:
		return dir + "/" + idMakerFile
	}
}
func (im *IdMaker) Record() {
	id := im.SeqId.id
	fmt.Println("------------------------------------------------------------------------------------")
	fmt.Println(fmt.Sprintf("program execution done, record id : %d, time : %s", id, time.Now().String()))
	fmt.Println("------------------------------------------------------------------------------------")
	ioutil.WriteFile(PathJoin(), []byte(strconv.Itoa(int(id))), 0664)
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
