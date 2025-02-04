package models

import (
	"context"
	"encoding/json"
	"fmt"
	"ginchat/utils"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	UserId     int64  //发送者
	TargetId   int64  //接受者
	Type       int    //发送类型 1私聊 2群聊 3广播
	Media      int    //消息类型 1文字 2表情包 3图片 4音频
	Content    string //消息内容
	CreateTime int64
	Pic        string
	Url        string
	Desc       string
	Amount     int //其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

type Node struct {
	Conn      *websocket.Conn
	DataQuene chan []byte
	GroupSets set.Interface
}

// 映射关系
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

// 需要： 发送者id， 接收者id， 发送类型， 消息类型， 消息内容
func Chat(writer http.ResponseWriter, request *http.Request) {
	//1.检验token合法性
	query := request.URL.Query()
	id := query.Get("userId")
	userId, _ := strconv.ParseInt(id, 10, 64)
	// msgType := query.Get("type")
	// targetId := query.Get("targetId")
	// context := query.Get("context")
	isvalida := true //checkToken() 待开发
	conn, err := (&websocket.Upgrader{
		//token 校验
		CheckOrigin: func(r *http.Request) bool {
			return isvalida
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	//2.获取conn
	node := &Node{
		Conn:      conn,
		DataQuene: make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe),
	}
	//3.用户关系
	//4.userid跟node绑定并加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()
	//5.完成发送逻辑
	go SendProc(node)
	//6.完成接收逻辑
	go RecvProc(node)

	SendMsg(userId, []byte("欢迎进入聊天室"))
}

func SendProc(node *Node) {
	for {
		fmt.Println("enter SendProc ......")
		select {
		case data := <-node.DataQuene:
			fmt.Println("[ws] sendProc >>>> msg: ", string(data))
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}

		}
	}
}

func RecvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("[ws] RecvProc <<<<<", string(data))
		Dispatch(data)
		broadMsg(data)

	}
}

var udpspendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpspendChan <- data
}

func init() {
	fmt.Println("init goroutine")
	go udpSendProc()
	go udpRecvProc()

}

// 完成udp数据发送协程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(192, 168, 10, 255),
		Port: 3000,
	})
	defer con.Close()
	if err != nil {
		fmt.Println(err)
	}
	for {
		select {
		case data := <-udpspendChan:
			fmt.Println("udpSendProc data :", string(data))
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// 完成udo数据接收协程
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 3000,
	})
	defer con.Close()
	if err != nil {
		fmt.Println(err)
	}
	for {
		var buf [512]byte

		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("udpRecvProc data :", string(buf[0:n]))
		//避免消息重复
		//Dispatch(buf[0:n])
	}
}

// 后端调度逻辑处理
func Dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch msg.Type {
	case 1: //私信
		fmt.Println("Dispatch data :", string(data))
		SendMsg(msg.TargetId, data)
		// case 2://群发
		// 	SendGroupMsg()
		// case 3://广播
		// 	SendAllMsg()
		// case 4:
	}
}

func SendMsg(userId int64, msg []byte) {
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()

	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	//传进函数中的userId是message中的targetId
	targetIdStr := strconv.Itoa(int(userId))
	//解析msg后，将message中的userId赋值给userId
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime = time.Now().Unix()
	//查找缓存中是否存在此发送者信息
	r, err := utils.Red.Get(ctx, "online_"+userIdStr).Result()
	if err != nil {
		fmt.Println(err)
	}
	if r != "" {
		//发送信息给接收者
		if ok {
			fmt.Println("sendMsg >>> userID: ", userId, " msg:", string(msg))
			//将消息发送至页面
			node.DataQuene <- msg
		}
	}
	var key string
	//对targetId和userId进行排序，小的在前数值大的在后面作为key
	//userId代表msg里面的targetId，jaonMsg.userId代表msg里面的userId
	if userId > jsonMsg.UserId {
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}
	//倒序排序
	res, err := utils.Red.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	score := float64(cap(res)) + 1
	//添加数据至redis
	ress, e := utils.Red.ZAdd(ctx, key, &redis.Z{score, msg}).Result()
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(ress)

}

// 获取缓存里面的消息
func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) []string {
	rwLocker.RLock()
	_, ok := clientMap[userIdA]
	rwLocker.RUnlock()
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}
	var rels []string
	var err error
	if isRev {
		//倒序
		rels, err = utils.Red.ZRevRange(ctx, key, start, end).Result()
	} else {
		rels, err = utils.Red.ZRange(ctx, key, start, end).Result()
	}

	if err != nil {
		fmt.Println(err)
	}
	if ok {
		return rels
		// for _, val := range rels {
		// 	//后台通过websocket将redis消息推送至页面

		// 	fmt.Println("sendMsg >>> userID: ", userIdA, " msg:", val)
		// 	node.DataQuene <- []byte(val)

		// }
	} else {
		//登录的时候会将userId存到前端session
		//未登录前端userId()方法无法获得userId
		//消息传递的时候无法添加userId至clientMap
		//后续node就未赋值
		//未读取到userId，即未登录
		fmt.Println("未登录,无法读取redis")
		return nil
	}

}
