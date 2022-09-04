package lib

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
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
			handleMessage(update, ob, bot)
			handleRequest(update, ob)
		}
	}(ob1)
}

func handleMessage(update tgbotapi.Update, ob *libonebot.OneBot, bot *tgbotapi.BotAPI) {
	type MyMessageEvent struct {
		libonebot.MessageEvent
		UserID    string `json:"user_id"`
		GroupID   string `json:"group_id"`
		ChannelID string `json:"channel_id"`
		GuildID   string `json:"guild_id"`
		SubType   string `json:"sub_type"`
	}
	detailType := "private"

	if update.Message == nil {
		// 回调消息
		if update.CallbackQuery != nil {
			if update.CallbackQuery.From.ID != update.CallbackQuery.Message.Chat.ID {
				detailType = "group"
			}
			pushEvent(ob, &MyMessageEvent{
				MessageEvent: libonebot.MakeMessageEvent(time.Now(), detailType, strconv.FormatInt(update.CallbackQuery.Message.Chat.ID, 10)+"_"+strconv.Itoa(update.CallbackQuery.Message.MessageID), libonebot.Message{libonebot.TextSegment(update.CallbackQuery.Data)}, update.CallbackQuery.Data),
				UserID:       strconv.FormatInt(update.CallbackQuery.From.ID, 10),
				GroupID:      strconv.FormatInt(update.CallbackQuery.Message.Chat.ID, 10),
				GuildID:      strconv.FormatInt(update.CallbackQuery.Message.Chat.ID, 10),
				SubType:      "call_back",
			})
			return
		} else {
			return
		}
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
	if update.Message.Entities != nil {
		for _, entity := range update.Message.Entities {
			if entity.Type == "text_mention" {
				messages = append(messages, libonebot.MentionSegment(strconv.FormatInt(entity.User.ID, 10)))
			} else if entity.Type == "bot_command" {
				if strings.Contains(update.Message.Text, bot.Self.UserName) {
					messages = append(messages, libonebot.MentionSegment(strconv.FormatInt(bot.Self.ID, 10)))
					update.Message.Text = strings.ReplaceAll(update.Message.Text, "@"+bot.Self.UserName, "")
				}

			}
		}
	}
	if update.Message.Text != "" {
		messages = append(messages, libonebot.TextSegment(update.Message.Text))
	}
	if update.Message.Photo != nil {
		messages = append(messages, libonebot.ImageSegment(update.Message.Photo[0].FileID))
	}

	channelID := ""

	if update.Message.From.ID != update.Message.Chat.ID {
		detailType = "group"
		if update.Message.SenderChat != nil && update.Message.SenderChat.Type == "channel" {
			detailType = "channel"
			channelID = strconv.FormatInt(update.Message.SenderChat.ID, 10)
		}
	}

	pushEvent(ob, &MyMessageEvent{
		MessageEvent: libonebot.MakeMessageEvent(time.Now(), detailType, strconv.FormatInt(update.Message.Chat.ID, 10)+"_"+strconv.Itoa(update.Message.MessageID), messages, update.Message.Text),
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
