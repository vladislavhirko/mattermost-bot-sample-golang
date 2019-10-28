package server

import (
	"fmt"
	"github.com/xr9kayu/mattermost-bot-sample-golang/bot"
	"github.com/mattermost/mattermost-server/model"
	"net/http"
)

type RequestHandler struct {
	Bot *bot.Bot
}

func (h *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	postAIR := model.PostActionIntegrationRequestFromJson(r.Body)
	switch postAIR.Context["action"].(string) {
	case "help_menu":
		fmt.Printf("HI")
		wsEvent :=model.NewWebSocketEvent(
			model.WEBSOCKET_EVENT_POSTED,
			postAIR.TeamId,
			postAIR.ChannelId,
			postAIR.UserId,
			map[string]bool{},
		)
		wsEvent.Data["channel_type"]=model.CHANNEL_DIRECT
		wsEvent.Data["msg_type"]="help_menu"
		h.Bot.ChannelHTTP <- wsEvent
		w.Write((&model.PostActionIntegrationResponse{}).ToJson())
	case "wellcome_menu":
		attach := h.Bot.Attachments["help_menu"]
		post := model.PostActionIntegrationResponse{
			Update: &model.Post{
				UserId:    postAIR.UserId,
				ChannelId: postAIR.ChannelId,
				RootId:    postAIR.PostId,
				Props: attach,
			},
		}
		w.Write(post.ToJson())
	}
}
