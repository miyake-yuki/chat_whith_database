package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

//roomはチャットルームを表す構造体
type room struct {
	forward chan *Message    //forwardは他のクライアントに転送するメッセージ用のチャネル
	join    chan *client     //joinはチャットルームに参加しようとするクライアントのためのチャネル
	leave   chan *client     //leaveはチャットルームから退出しようとするクライアントのためのチャネル
	clients map[*client]bool //clientsは在室している全てのクライアントが保持されるマップ
}

func newRoom() *room {
	return &room{
		forward: make(chan *Message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

//このServeHTTPは /room へのリクエストに対応して、
//websocket通信によるチャットを開始する

//TODO:ServeHTTPがいつ終了するかを調べる->readの終了時=ReadJSONでエラーが発生した時
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//通信をWebSocketへアップグレード
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
		return
	}
	//クライアントを生成
	client := &client{
		name:   "",
		socket: socket,
		send:   make(chan *Message, messageBufferSize),
		room:   r,
	}
	//クライアントをroomに入室させる
	r.join <- client
	//このAPIが終了する時にクライアントを退室させる
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}

func (r *room) run() {
	// テーブルに接続
	table := connectDB().Table(tableName)
	for {
		select {
		//入室
		case client := <-r.join:
			r.clients[client] = true
		//退室
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			//メッセージを受信
		case mess := <-r.forward:
			//データベースに登録
			if err := table.Put(newData(mess)).Run(); err != nil {
				fmt.Println(err)
			}
			//全てのクライアントに送信
			for client := range r.clients {
				select {
				//メッセージを送信
				case client.send <- mess:
				//送信に失敗
				default:
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}
