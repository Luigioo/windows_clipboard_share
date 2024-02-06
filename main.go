package main

import (
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/go-toast/toast"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

const (
	LocalPort  = "3000"                     // 本地监听端口
	RemoteAddr = "http://192.168.80.4:3001" // 远程设备的 IP 地址和端口
)

var (
	modifiedContent atomic.Value
)

func sendData(data string) {
	fmt.Println("发送数据:", data)
	resp, err := http.Post(RemoteAddr, "text/plain", strings.NewReader(data))
	if err != nil {
		fmt.Println("发送数据失败:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("数据发送成功，状态码:", resp.StatusCode)
}

func monitorClipboard() {
	var lastValue string
	fmt.Println("剪贴板监控就绪，正在发送数据...")

	for {
		currentValue, err := clipboard.ReadAll()
		if err != nil {
			fmt.Println("无法读取剪贴板:", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if currentValue != lastValue && currentValue != modifiedContent.Load().(string) {
			sendData(currentValue)
			lastValue = currentValue
		}
		time.Sleep(1 * time.Second)
	}
}
func showNotification(content string) {
	// 对过长的内容进行裁剪
	displayContent := content
	if len(content) > 100 {
		displayContent = content[:100] + "..."
	}

	// 获取当前时间
	timestamp := time.Now().Format("15:04:05")

	notification := toast.Notification{
		AppID:   "Go Clipboard Sync",
		Title:   "剪贴板更新",
		Message: fmt.Sprintf("接收到新内容 (%s):\n%s", timestamp, displayContent),
	}
	err := notification.Push()
	if err != nil {
		fmt.Println("无法发送通知:", err)
	}
}

func handleRequests() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
			content := string(body)
			fmt.Println("接收到数据:", content)

			err = clipboard.WriteAll(content)
			if err != nil {
				fmt.Println("无法写回剪贴板:", err)
			} else {
				modifiedContent.Store(content)
				showNotification(content)
			}

			fmt.Fprintln(w, "数据接收成功")
		}
	})

	fmt.Println("HTTP 服务器启动，正在监听端口:", LocalPort)
	http.ListenAndServe(":"+LocalPort, nil)
}

func main() {
	modifiedContent.Store("")

	go monitorClipboard()
	go handleRequests()

	select {} // 防止主 goroutine 直接退出
}
