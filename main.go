package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

const tableName = "websocket_test"

//rootHandlerは、テンプレートから作成され
//http.Handlerインターフェースを満たすハンドラー
//ルートディレクトリへのリクエストに対応する
type rootHandler struct {
	once     sync.Once          // 一度だけ実行される様にするための変数
	filename string             // テンプレートファイルの名前
	temp     *template.Template // テンプレートへの参照
}

func (root *rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//テンプレートを一度だけパースする
	root.once.Do(func() {
		root.temp = template.Must(template.ParseFiles(filepath.Join("templates", root.filename)))
	})
	//テンプレートを表示（wに流し込む）
	root.temp.Execute(w, nil)
}

func main() {
	r := newRoom()
	//ルートにハンドラーを登録
	http.Handle("/", &rootHandler{filename: "index.html"})
	// /chatにハンドラを登録
	http.Handle("/chat", &chatHandler{filename: "chat.html"})
	// /roomにハンドラを登録
	http.Handle("/room", r)
	//チャットルームを開始
	go r.run()
	// webサーバを開始
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
