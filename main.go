package main
import (
	b "github.com/xr9kayu/mattermost-bot-sample-golang/bot"
	s "github.com/xr9kayu/mattermost-bot-sample-golang/server"
	a "github.com/xr9kayu/mattermost-bot-sample-golang/attachments"
)

func main(){
	bot := b.NewBot("http://localhost:8065", a.Parse())
	go bot.Start()
	server := s.NewServer(":8090")
	server.RegisterHandleFunc("/",&s.RequestHandler{Bot:bot})
	server.Start()
}
