package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/guregu/dynamo"
)

type chatHandler struct {
	filename  string             //テンプレートファイルの名前
	loginData *Login             //ログイン時に送られてくるデータ
	chatTemp  *template.Template //テンプレートへの参照
	log       LogAndName         //ユーザの名前と過去の会話履歴
}

// Login はユーザがログインする時に送ってくる
// JSONをパースする用の構造体
type Login struct {
	RoomID uint64 `json:"room_id" dynamo:"room_id"`
	Name   string `json:"username" dynamo:"name"`
}

// LogAndName 構造体はテンプレートにクラアントの名前と
// チャットのログを10件分表示するための構造体
type LogAndName struct {
	PastMessages []Data
	Name         string
}

func newChatHandler(filename string) *chatHandler {
	var loginData Login
	return &chatHandler{
		filename:  filename,
		loginData: &loginData,
	}
}

func (chat *chatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//POSTされてきたBodyの長さを取得
	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		fmt.Println("Error in Atoi:", err)
		return
	}
	//Body(ユーザ名とチャットルームの番号)を読み取る
	body := make([]byte, length)
	if _, err := r.Body.Read(body); err != nil && err != io.EOF {
		fmt.Println("Error in Read:", err)
		return
	}
	//BodyのJSONをloginDataに入れる
	if err := json.Unmarshal(body, chat.loginData); err != nil {
		fmt.Println("Error in Unmarshal:", err)
		return
	}
	//ユーザ名とチャットルームの番号をwhitelist_testに問い合わせる
	var result Login
	table := connectDB().Table("whitelist_test")
	fmt.Println("DynamoDBに接続")
	if err := table.Get("room_id", chat.loginData.RoomID).
		Range("name", dynamo.Equal, chat.loginData.Name).
		One(&result); err != nil {
		//whitelist_testに登録されていなかった場合
		if err == dynamo.ErrNotFound {
			fmt.Println("whitelistに弾かれた")
			w.Write(body)
			//よくわからないエラー
		} else {
			fmt.Println("Error from One:", err)
			return
		}
		//きちんとwhitelist_testに乗っていた場合
	} else {
		//クライアントを生成
		client := newClient(chat.loginData.Name, chat.loginData.RoomID)
		//このクライアントのWebSocket通信用のAPIを作成
		http.Handle("/"+string(client.hash), client)
		//クライアント側にWebSocket通信に使用するAPIのpathを伝達するために、
		//APIのpathをCookieに焼く
		http.SetCookie(w, &http.Cookie{
			Value: string(client.hash),
		})
		//過去10件のデータを取得
		log := getLast10Data(chat.loginData.RoomID)
		//htmlに埋め込むデータを生成
		userData := &LogAndName{PastMessages: log, Name: chat.loginData.Name}
		// テンプレートをパース
		chat.chatTemp = template.Must(template.ParseFiles(filepath.Join("templates", chat.filename)))
		// テンプレートにログと名前を埋め込んで表示
		chat.chatTemp.Execute(w, userData)
	}
}
