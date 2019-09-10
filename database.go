package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

// Data 構造体はdynamoDBから受け取ったデータを入れる構造体
type Data struct {
	RoomID  uint64 `dynamo:"room_id"`
	Name    string `dynamo:"name"`
	Message string `dynamo:"message"`
	Time    int64  `dynamo:"timestamp"`
}

func newData(mes *Message) *Data {
	return &Data{
		RoomID:  0,
		Name:    mes.Content.Name,
		Message: mes.Content.Message,
		Time:    mes.Content.When,
	}
}

func connectDB() *dynamo.DB {
	// ~/.aws/credentialsの[test]から認証情報を取ってくる
	cred := credentials.NewSharedCredentials("", "test")

	return dynamo.New(session.New(), &aws.Config{
		Credentials: cred,
		Region:      aws.String("us-east-2"),
	})
}

func getLast10Data(id uint64) []Data {
	var last10data []Data
	// DBに接続
	table := connectDB().Table("websocket_test")
	// 最新10件を取得
	table.Get("room_id", id).Order(dynamo.Descending).Limit(10).All(&last10data)

	return last10data
}
