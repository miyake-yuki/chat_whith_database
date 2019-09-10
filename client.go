package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

//client構造体はチャットを行う一人のユーザを表す
type client struct {
	hash   uint64          //hashはnameとroomIDから生成されるハッシュ値
	name   string          //nameはこのユーザの名前
	socket *websocket.Conn //socketはこのクライアント用のWebSocketへの参照
	send   chan *Message   //sendはメッセージが送られてくるチャネル
	room   *room           //roomはこのクライアントが参加しているチャットルームへの参照
}

//Message 構造体は送信されたメッセージを表す
type Message struct {
	FirstTime bool  `json:"FirstTime"` //FirstTimeは初回の名前登録用の通信か、通常のメッセージの通信かの真偽値
	Content   *Body `json:"Content"`   //Contentはメッセージ本体
}

// Body 構造体はメッセージの本体を表す
type Body struct {
	Name    string `json:"Name"`    //Nameは発言者の名前
	Message string `json:"Message"` //Messageはメッセージの内容
	When    int64  //Whenは発言した時間(UNIX NANO TIME)
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

func newClient(name string, roomID uint64) *client {
	//roomIDを[]byteに変換
	buffer := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(buffer, roomID)
	//nameとroomIDのバイト列を結合
	buffer = append(buffer, []byte(name)...)
	//bufferからハッシュ値を生成
	hash := sha1.Sum(buffer)
	//ハッシュ値をuint64に変換
	value, _ := binary.Uvarint(hash[:])
	//クライアントを生成
	client := &client{
		hash: value,
		name: name,
		send: make(chan *Message, messageBufferSize),
		room: globalApart.offerRoom(roomID),
	}
	//クライアントを入室させる
	client.room.join <- client

	return client
}

func (c *client) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//通信をWebSocketへアップグレード
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
		return
	}
	c.socket = socket
	//このメソッドの終了時にクライアントを退室させる
	defer func() { c.room.leave <- c }()
	//クライアントの読み書きを開始する
	go c.write()
	c.read()
}

//read()は、clientがWebSocketを通じてサーバに送信したデータをroomへ送る
func (c *client) read() {
	var mess *Message
	for {
		if err := c.socket.ReadJSON(&mess); err == nil {
			if mess.FirstTime {
				//WebSocket通信が開始された最初の通信(onopen)で、クライアントに名前を登録
				c.name = mess.Content.Name
			} else {
				//通常のメッセージの通信については、NameとWhenを追記してroomに送信
				mess.Content.Name = c.name
				mess.Content.When = time.Now().UnixNano()
				c.room.forward <- mess
			}
		} else {
			fmt.Println("Error in read():", err)
			break
		}
	}
	c.socket.Close() //クライアントのWebSocket通信を切断
}

//write()は、クライアントに送られてきたメッセージをチャネルから取り出して
//JSONにパースし、WebSocketを通じてクライアントのブラウザ上に書き込む
func (c *client) write() {
	for mess := range c.send { //チャネルが配列ならこの様に情報を取り出すのが普通らしい
		if err := c.socket.WriteJSON(mess); err != nil {
			fmt.Println("Error in write():", err)
			break
		}
	}
	c.socket.Close() //クライアントのWebSocket通信を切断
}
