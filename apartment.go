package main

type apartment struct {
	build     chan *room      //buildは新しく生成されたroomを登録するためのチャネル
	demolish  chan *room      //demolishは誰もいなくなったroomを消去するためのチャネル
	rooms     map[*room]bool  //roomsは現在稼働しているroomを管理するためのチャネル
	moveIn    chan *client    //moveInはチャットに参加してきた人を登録するためのチャネル
	moveOut   chan *client    //moveOutはチャットをやめた人を消去するためのチャネル
	residents map[int]*client //residentsはチャットに参加している人を管理するためのマップ
}

func newApartment() *apartment {
	return &apartment{
		build:     make(chan *room),
		demolish:  make(chan *room),
		rooms:     make(map[*room]bool),
		moveIn:    make(chan *client),
		moveOut:   make(chan *client),
		residents: make(map[int]*client),
	}
}

//manageはapart内のroomを管理する
func (apart apartment) manage() {
	for {
		select {
		case room := <-apart.build:
			//新しいroomが生成された場合
			apart.rooms[room] = true
			//チャットルームを開始
			go room.run()
		case room := <-apart.demolish:
			//roomが消去される場合
			delete(apart.rooms, room)
		}
	}
}

//offerRoomはidに応じたroomを提供し、なければ新たにroomを生成しapartに登録する
func (apart apartment) offerRoom(id uint64) *room {
	for room := range apart.rooms {
		if room.id == id {
			return room
		}
	}
	brandNewRoom := newRoom(id)
	apart.build <- brandNewRoom
	return brandNewRoom
}
