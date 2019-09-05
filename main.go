package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

//templateHanlerは、テンプレートから作成され
//http.Handlerインターフェースを満たすハンドラー
type templateHandler struct {
	once     sync.Once          //一度だけ実行される様にするための変数
	filename string             //テンプレートファイルの名前
	temp1    *template.Template // テンプレートへの参照
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//テンプレートを一度だけパースする
	t.once.Do(func() {
		t.temp1 = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	//テンプレートを表示（wに流し込む）
	t.temp1.Execute(w, nil)
}

func main() {
	r := newRoom()
	//ルートにハンドラーを登録
	http.Handle("/", &templateHandler{filename: "chat.html"})
	// /roomにハンドラを登録
	http.Handle("/room", r)
	//チャットルームを開始
	go r.run()
	// webサーバを開始
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
