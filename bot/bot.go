package bot

import (
	"fmt"
	"github.com/mattermost/mattermost-server/model"
	logger "github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

const (
	SAMPLE_NAME      = "Mattermost Bot Sample"
	USER_EMAIL       = "bot@example.com"
	USER_PASSWORD    = "password1"
	USER_NAME        = "samplebot"
	TEAM_NAME        = "botsample"
	CHANNEL_LOG_NAME = "debugging-for-sample-bot"
)

type Bot struct {
	// mattermost structures
	botUser *model.User
	botTeam *model.Team
	// http websocket clients
	httpClient    *model.Client4
	wsocketClient *model.WebSocketClient
	// logger instance
	Logger *logger.Entry
	ChannelHTTP chan *model.WebSocketEvent
	// attachments to create interactive posts
	Attachments map[string]map[string]interface{}
}

func NewBot(mattermostURL string,
	props map[string]map[string]interface{}) *Bot {
	log := logger.New()
	log.SetLevel(logger.DebugLevel)
	log.SetOutput(os.Stdout)
	entry := log.WithFields(logger.Fields{
		"package": "bot",
	})
	return &Bot{
		botUser:     new(model.User),
		botTeam:     new(model.Team),
		httpClient:  model.NewAPIv4Client(mattermostURL),
		Attachments: props,
		Logger:      entry,
		ChannelHTTP: make(chan *model.WebSocketEvent, 1),
	}
}

func (b *Bot) Start() {
	b.setupGracefulShutdown()
	b.makeSureServerIsRunning()
	if err := b.login(); err != nil {
		b.Logger.Fatalln(err.Error())
	}
	b.findBotTeam()
	var err *model.AppError
	b.wsocketClient, err = model.NewWebSocketClient4("ws://localhost:8065", b.httpClient.AuthToken)
	if err != nil {
		b.Logger.Fatalln(err.Error())
	}
	b.wsocketClient.Listen()
	for {
		select {
		case resp := <-b.ChannelHTTP:
			fmt.Println("HDEUHDEUDUH")
			b.handleWebSocketResponse(resp)
		case resp := <-b.wsocketClient.EventChannel:
			b.handleWebSocketResponse(resp)
		}
	}
}

func (b *Bot) login() error {
	if user, resp := b.httpClient.Login(USER_NAME, USER_PASSWORD); resp.Error != nil {
		b.Logger.Fatal(resp.Error.Error())
	} else {
		b.botUser = user
		fmt.Println(user.Username)
	}
	return nil
}

func (b *Bot) makeSureServerIsRunning() {
	if props, resp := b.httpClient.GetOldClientConfig(""); resp.Error != nil {
		println("There was a problem pinging the Mattermost server.  Are you sure it's running?")
		b.Logger.Fatalln(resp.Error.Error())
	} else {
		println("Server detected and is running version " + props["Version"])
	}
}

func (b *Bot) findBotTeam() {
	if team, resp := b.httpClient.GetTeamByName(TEAM_NAME, ""); resp.Error != nil {
		b.Logger.Fatalln(resp.Error.Error())
	} else {
		b.botTeam = team
	}
}

func (b *Bot) sendMsgToChannel(msg, replyToId, channelId string, msgType string) {
	post := &model.Post{}
	post.ChannelId = channelId
	//if msgType != "" {
	//	if props, ok := b.propsMap[msgType]; ok {
	//		post.Props = props
	//	}
	//}
	mapa := b.Attachments[msgType]
	post.Props = mapa
	if _, resp := b.httpClient.CreatePost(post); resp.Error != nil {
		b.Logger.Error(resp.Error.Error())
	}
}

func (b *Bot) handleWebSocketResponse(event *model.WebSocketEvent) {
	switch event.Event {
	case model.WEBSOCKET_EVENT_POSTED:
		switch event.Data["channel_type"] {
		case model.CHANNEL_DIRECT:
			switch event.Data["channel_display_name"] {
			case "@" + b.botUser.Username:
				return
			default:
				b.handleMsgFromDirectChannel(event)
			}
		case model.CHANNEL_OPEN:
			return
		}
	case model.WEBSOCKET_EVENT_NEW_USER:
		b.handleNewUser(event)
	}
}

func (b *Bot) handleNewUser(event *model.WebSocketEvent) {
	// Create direct channel for new user
	userId, _ := event.Data["user_id"].(string)
	channel, resp := b.httpClient.CreateDirectChannel(b.botUser.Id, userId)
	if resp.Error != nil {
		b.Logger.Errorln(resp.Error.Error())
	}
	b.sendMsgToChannel("Hello", "", channel.Id, "wellcome_menu")
}

func (b *Bot) handleMsgFromDirectChannel(event *model.WebSocketEvent) {
	channelId := event.Broadcast.ChannelId
	// Lets only reponded to messaged posted events
	if event.Event != model.WEBSOCKET_EVENT_POSTED {
		return
	}

	b.Logger.Info(event.Data)
	msgtype, ok := event.Data["msg_type"].(string)
	if !ok {
		msgtype = "wellcome_menu"
	}
	b.sendMsgToChannel("", "", channelId, msgtype)
}

func (b *Bot) setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			if b.wsocketClient != nil {
				b.wsocketClient.Close()
			}
			os.Exit(0)
		}
	}()
}

//func HandleMsgFromDebuggingChannel(event *model.WebSocketEvent) {
//	// If this isn't the debugging channel then lets ingore it
//	if event.Broadcast.ChannelId != debuggingChannel.Id {
//		return
//	}
//
//	// Lets only reponded to messaged posted events
//	if event.Event != model.WEBSOCKET_EVENT_POSTED {
//		return
//	}
//
//	println("responding to debugging channel msg")
//
//	post := model.PostFromJson(strings.NewReader(event.Data["post"].(string)))
//
//	if post != nil {
//
//		// ignore my events
//		if post.UserId == botUser.Id {
//			return
//		}
//
//		// if you see any word matching 'alive' then respond
//		if matched, _ := regexp.MatchString(`(?:^|\W)alive(?:$|\W)`, post.Message); matched {
//			SendMsgToChannel("Yes I'm running", post.Id, debuggingChannel.Id)
//			return
//		}
//
//		// if you see any word matching 'up' then respond
//		if matched, _ := regexp.MatchString(`(?:^|\W)up(?:$|\W)`, post.Message); matched {
//			SendMsgToChannel("Yes I'm running", post.Id, debuggingChannel.Id)
//			return
//		}
//
//		// if you see any word matching 'running' then respond
//		if matched, _ := regexp.MatchString(`(?:^|\W)running(?:$|\W)`, post.Message); matched {
//			SendMsgToChannel("Yes I'm running", post.Id, debuggingChannel.Id)
//			return
//		}
//
//		// if you see any word matching 'hello' then respond
//		if matched, _ := regexp.MatchString(`(?:^|\W)hello(?:$|\W)`, post.Message); matched {
//			SendMsgToChannel("Yes I'm running", post.Id, debuggingChannel.Id)
//			return
//		}
//	}
//
//	SendMsgToChannel("I did not understand you!", post.Id, debuggingChannel.Id)
//}
//func CreateBotDebuggingChannelIfNeeded() {
//	if rchannel, resp := client.GetChannelByName(CHANNEL_LOG_NAME, botTeam.Id, ""); resp.Error != nil {
//		println("We failed to get the channels")
//		PrintError(resp.Error)
//	} else {
//		debuggingChannel = rchannel
//		valueOfDebugingChanel := reflect.ValueOf(debuggingChannel)
//		typeOfDebugingChanel := reflect.TypeOf(debuggingChannel)
//		if valueOfDebugingChanel.Kind() == reflect.Ptr {
//			valueOfDebugingChanel = valueOfDebugingChanel.Elem()
//		}
//		if typeOfDebugingChanel.Kind() == reflect.Ptr {
//			typeOfDebugingChanel = typeOfDebugingChanel.Elem()
//		}
//
//		for i := 0; i < valueOfDebugingChanel.NumField(); i++ {
//			fmt.Println(typeOfDebugingChanel.Field(i).Name, ": ", valueOfDebugingChanel.Field(i))
//		}
//		return
//	}
//
//	// Looks like we need to create the logging channel
//	channel := &model.Channel{}
//	channel.Name = CHANNEL_LOG_NAME
//	channel.DisplayName = "Debugging For Sample Bot"
//	channel.Purpose = "This is used as a test channel for logging bot debug messages"
//	channel.Type = model.CHANNEL_OPEN
//	channel.TeamId = botTeam.Id
//	if rchannel, resp := client.CreateChannel(channel); resp.Error != nil {
//		println("We failed to create the channel " + CHANNEL_LOG_NAME)
//		PrintError(resp.Error)
//	} else {
//		debuggingChannel = rchannel
//		println("Looks like this might be the first run so we've created the channel " + CHANNEL_LOG_NAME)
//	}
//}
