package main

import "github.com/gorilla/websocket"

// client構造体はチャットを行う一人のユーザを表す
type client struct {
	socket *websocket.Conn //socketはこのクライアント用のWebSocket
	send   chan []byte     //sendはメッセージが送られてくるチャネル
	room   *room           //roomはこのクライアントが参加しているチャットルームへの参照
}

//read()は、clientがブラウザ上で書き込んだ情報を読み取り
//WebSocketを通じてサーバに送信し、room構造体のfowardチャネルに送る
func (c *client) read() {
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close() //クライアントのWebSocket通信を切断
}

//write()は、クライアントに送られてきたメッセージをチャネルから取り出して
//WebSocketを通じてクライアントのブラウザ上に書き込む
func (c *client) write() {
	for msg := range c.send { //チャネルが配列ならこの様に情報を取り出すのが普通らしい
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close() //クライアントのWebSocket通信を切断
}
