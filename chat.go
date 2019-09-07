package main

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type chatHandler struct {
	filename string             //テンプレートファイルの名前
	chatTemp *template.Template // テンプレートへの参照
}

// LogAndName 構造体はテンプレートにクラアントの名前と
// チャットのログを10件分表示するための構造体
type LogAndName struct {
	PastMessages []Data
	Name         string
}

func (chat *chatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 過去10件のデータを取得
	log := getLast10Data()
	// クライアントの名前を取得
	clientName := r.FormValue("username")
	// htmlに埋め込むデータを生成
	userData := &LogAndName{PastMessages: log, Name: clientName}
	// テンプレートをパース
	chat.chatTemp = template.Must(template.ParseFiles(filepath.Join("templates", chat.filename)))
	// テンプレートにログと名前を埋め込んで表示
	chat.chatTemp.Execute(w, userData)

}
