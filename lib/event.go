package lib

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/botuniverse/go-libonebot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleEvent(ob1 *libonebot.OneBot, bot *tgbotapi.BotAPI) {
	channel := bot.GetUpdatesChan(tgbotapi.NewUpdate(1))
	go func(ob *libonebot.OneBot) {
		for true {
			update := <-channel
			data, _ := json.Marshal(update)
			_ = os.WriteFile("update.json", data, 0666)
			handleMessage(update, ob)
			handleRequest(update, ob)
		}
	}(ob1)
}

func handleMessage(update tgbotapi.Update, ob *libonebot.OneBot) {
	if update.Message == nil {
		return
	}
	if update.Message.Text == "" && update.Message.Photo == nil {
		return
	}
	var messages libonebot.Message
	if update.Message.ReplyToMessage != nil {
		messages = append(messages,
			libonebot.ReplySegment(strconv.FormatInt(update.Message.Chat.ID, 10)+"_"+strconv.Itoa(update.Message.MessageID),
				strconv.FormatInt(update.Message.From.ID, 10)))
	}
	if update.Message.Text != "" {
		messages = append(messages, libonebot.TextSegment(update.Message.Text))
	}
	if update.Message.Photo != nil {
		messages = append(messages, libonebot.ImageSegment(update.Message.Photo[0].FileID))
	}
	channelID := ""

	detail_type := "private"
	if update.Message.From.ID != update.Message.Chat.ID {
		detail_type = "group"
		if update.Message.SenderChat != nil && update.Message.SenderChat.Type == "channel" {
			detail_type = "channel"
			channelID = strconv.FormatInt(update.Message.SenderChat.ID, 10)
		}
	}

	type MyMessageEvent struct {
		libonebot.MessageEvent
		UserID    string `json:"user_id"`
		GroupID   string `json:"group_id"`
		ChannelID string `json:"channel_id"`
		GuildID   string `json:"guild_id"`
	}

	pushEvent(ob, &MyMessageEvent{
		MessageEvent: libonebot.MakeMessageEvent(time.Now(), detail_type, strconv.FormatInt(update.Message.Chat.ID, 10)+"_"+strconv.Itoa(update.Message.MessageID), messages, update.Message.Text),
		UserID:       strconv.FormatInt(update.Message.From.ID, 10),
		GroupID:      strconv.FormatInt(update.Message.Chat.ID, 10),
		GuildID:      strconv.FormatInt(update.Message.Chat.ID, 10),
		ChannelID:    channelID,
	})

}

func handleRequest(update tgbotapi.Update, ob *libonebot.OneBot) {
	if update.Message != nil && update.Message.LeftChatMember != nil {
		subType := "leave"
		if update.Message.From.ID != update.Message.LeftChatMember.ID {
			subType = "kick"
		}
		pushEvent(ob, &struct {
			libonebot.NoticeEvent
			GroupID    string `json:"group_id"`
			UserID     string `json:"user_id"`
			OperatorID string `json:"operator_id"`
		}{
			NoticeEvent: libonebot.NoticeEvent{Event: libonebot.Event{
				ID:         strconv.Itoa(update.UpdateID),
				Time:       float64(time.Now().Unix()),
				Type:       libonebot.EventTypeNotice,
				DetailType: "group_member_decrease",
				SubType:    subType,
			},
			},
			GroupID:    strconv.FormatInt(update.Message.Chat.ID, 10),
			UserID:     strconv.FormatInt(update.Message.LeftChatMember.ID, 10),
			OperatorID: strconv.FormatInt(update.Message.From.ID, 10),
		})
	}
	if update.Message != nil && update.Message.NewChatMembers != nil {
		subType := ""

		for _, member := range update.Message.NewChatMembers {
			if update.Message.From.ID != member.ID {
				subType = "invite"
			} else {
				subType = "join"
			}
			pushEvent(ob, &struct {
				libonebot.NoticeEvent
				GroupID    string `json:"group_id"`
				UserID     string `json:"user_id"`
				OperatorID string `json:"operator_id"`
			}{
				NoticeEvent: libonebot.NoticeEvent{Event: libonebot.Event{
					Time:       float64(time.Now().Unix()),
					Type:       "notice",
					DetailType: "group_member_increase",
					SubType:    subType,
				}},
				GroupID:    strconv.FormatInt(update.Message.Chat.ID, 10),
				UserID:     strconv.FormatInt(member.ID, 10),
				OperatorID: strconv.FormatInt(update.Message.From.ID, 10),
			})
		}
	}

}

func pushEvent(ob *libonebot.OneBot, event libonebot.AnyEvent) {
	for _, comm := range comms {
		comm(ob, event)
	}
	ob.Push(event.(libonebot.AnyEvent))
}
