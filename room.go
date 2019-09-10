package main

import (
	"fmt"
)

//roomはチャットルームを表す構造体
type room struct {
	id      uint64
	forward chan *Message    //forwardは他のクライアントに転送するメッセージ用のチャネル
	join    chan *client     //joinはチャットルームに参加しようとするクライアントのためのチャネル
	leave   chan *client     //leaveはチャットルームから退出しようとするクライアントのためのチャネル
	clients map[*client]bool //clientsは在室している全てのクライアントが保持されるマップ
}

func newRoom(id uint64) *room {
	return &room{
		id:      id,
		forward: make(chan *Message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	// テーブルに接続
	table := connectDB().Table("websocket_test")
	for {
		select {
		//入室
		case client := <-r.join:
			r.clients[client] = true
		//退室
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			//もしクライアントがいなくなった場合、このgoroutineを終了する
			if len(r.clients) == 0 {
				globalApart.demolish <- r
				return
			}
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
