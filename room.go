package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

//roomはチャットルームを表す構造体
type room struct {
	forward chan []byte      //forwardは他のクライアントに転送するメッセージ用のチャネル
	join    chan *client     //joinはチャットルームに参加しようとするクライアントのためのチャネル
	leave   chan *client     //leaveはチャットルームから退出しようとするクライアントのためのチャネル
	clients map[*client]bool //clientsは在室している全てのクライアントが保持されるマップ
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
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

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
		return
	}

	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}

	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}

func (r *room) run() {
	for {
		select {
		//入室
		case client := <-r.join:
			r.clients[client] = true
		//退屋
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
		//全てのクライアントにメッセージを転送
		case msg := <-r.forward:
			for client := range r.clients {
				select {
				//メッセージを送信
				case client.send <- msg:
				//送信に失敗
				default:
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}
