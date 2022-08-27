package main

import (
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
			handleMessage(update, ob)
		}
	}(ob1)
}

func handleMessage(update tgbotapi.Update, ob *libonebot.OneBot) {
	if update.Message == nil {
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

	detail_type := "private"
	if update.Message.From.ID != update.Message.Chat.ID {
		detail_type = "group"
	}

	ob.Push(&libonebot.MessageEvent{
		Event:      libonebot.Event{ID: strconv.Itoa(update.UpdateID), Type: libonebot.EventTypeMessage, DetailType: detail_type, Time: float64(time.Now().Unix())},
		MessageID:  strconv.FormatInt(update.Message.Chat.ID, 10) + "_" + strconv.Itoa(update.Message.MessageID),
		Message:    messages,
		AltMessage: update.Message.Text,
	})

}
