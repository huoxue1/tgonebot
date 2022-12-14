package utils

import (
	"encoding/base64"
	"os"
	"strconv"
	"strings"

	"github.com/botuniverse/go-libonebot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

func MessageToChattables(bot *tgbotapi.BotAPI, msg libonebot.Message, chatId int64) []tgbotapi.Chattable {

	var results []tgbotapi.Chattable
	var replyId int

	var mention string

	for _, segment := range msg {
		switch segment.Type {
		case libonebot.SegTypeText:
			{
				text, err := segment.Data.GetString("text")
				if err != nil {
					log.Errorln("错误的消息段，已忽略")
					continue
				}
				if len(text) > 4096 {
					message := tgbotapi.NewMessage(chatId, "长文本消息，已转为文件发送")
					message.ReplyToMessageID = replyId
					results = append(results, message)
					file := tgbotapi.NewDocument(chatId, tgbotapi.FileBytes{
						Name:  "长文本.txt",
						Bytes: []byte(text),
					})
					file.ReplyToMessageID = replyId
					results = append(results, file)
				} else {
					message := tgbotapi.NewMessage(chatId, text)
					message.ReplyToMessageID = replyId
					results = append(results, message)
				}
			}

		case libonebot.SegTypeImage:
			{
				fileID, _ := segment.Data.GetString("file_id")
				var photo tgbotapi.PhotoConfig
				if strings.HasPrefix(fileID, "base64://") {
					data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(fileID, "base64://"))
					if err != nil {
						continue
					}
					os.WriteFile("qrcode.jpg", data, 0666)
					photo = tgbotapi.NewPhoto(chatId, tgbotapi.FileBytes{
						Name:  "file",
						Bytes: data,
					})

				} else if strings.HasPrefix(fileID, "file:///") {

					photo = tgbotapi.NewPhoto(chatId, tgbotapi.FilePath(strings.TrimPrefix(fileID, "file:///")))

				} else if strings.HasPrefix(fileID, "http://") || strings.HasPrefix(fileID, "https://") {
					photo = tgbotapi.NewPhoto(chatId, tgbotapi.FileURL(fileID))

				} else {
					photo = tgbotapi.NewPhoto(chatId, tgbotapi.FileID(fileID))
				}
				photo.ReplyToMessageID = replyId
				results = append(results, photo)
			}

		case libonebot.SegTypeAudio:
			var audio tgbotapi.AudioConfig
			{
				fileID, _ := segment.Data.GetString("file_id")
				if strings.HasPrefix(fileID, "base64://") {
					data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(fileID, "base64://"))
					if err != nil {
						continue
					}

					audio = tgbotapi.NewAudio(chatId, tgbotapi.FileBytes{
						Name:  "file",
						Bytes: data,
					})
				} else if strings.HasPrefix(fileID, "file:///") {
					audio = tgbotapi.NewAudio(chatId, tgbotapi.FilePath(strings.TrimPrefix(fileID, "file:///")))

				} else if strings.HasPrefix(fileID, "http://") || strings.HasPrefix(fileID, "https://") {
					audio = tgbotapi.NewAudio(chatId, tgbotapi.FileURL(fileID))

				} else {
					audio = tgbotapi.NewAudio(chatId, tgbotapi.FileID(fileID))
				}
				audio.ReplyToMessageID = replyId
				results = append(results, audio)
			}
		case libonebot.SegTypeVideo:
			{

				fileID, _ := segment.Data.GetString("file_id")
				var video tgbotapi.VideoConfig
				if strings.HasPrefix(fileID, "base64://") {
					data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(fileID, "base64://"))
					if err != nil {
						continue
					}
					video = tgbotapi.NewVideo(chatId, tgbotapi.FileBytes{
						Name:  "file",
						Bytes: data,
					})

				} else if strings.HasPrefix(fileID, "file:///") {
					video = tgbotapi.NewVideo(chatId, tgbotapi.FilePath(strings.TrimPrefix(fileID, "file:///")))

				} else if strings.HasPrefix(fileID, "http://") || strings.HasPrefix(fileID, "https://") {
					video = tgbotapi.NewVideo(chatId, tgbotapi.FileURL(fileID))
				} else {
					video = tgbotapi.NewVideo(chatId, tgbotapi.FileID(fileID))
				}
				video.ReplyToMessageID = replyId
				results = append(results, video)

			}
		case libonebot.SegTypeFile:
			{
				fileID, _ := segment.Data.GetString("file_id")
				var doc tgbotapi.DocumentConfig
				if strings.HasPrefix(fileID, "base64://") {
					data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(fileID, "base64://"))
					if err != nil {
						continue
					}
					doc = tgbotapi.NewDocument(chatId, tgbotapi.FileBytes{
						Name:  "document",
						Bytes: data,
					})
				} else if strings.HasPrefix(fileID, "file:///") {
					doc = tgbotapi.NewDocument(chatId, tgbotapi.FilePath(strings.TrimPrefix(fileID, "file:///")))

				} else if strings.HasPrefix(fileID, "http://") || strings.HasPrefix(fileID, "https://") {
					doc = tgbotapi.NewDocument(chatId, tgbotapi.FileURL(fileID))
				} else {
					doc = tgbotapi.NewDocument(chatId, tgbotapi.FileID(fileID))

				}
				doc.ReplyToMessageID = replyId
				results = append(results, doc)
			}
		case libonebot.SegTypeLocation:
			{
				latitude, _ := segment.Data.GetFloat64("latitude")
				longitude, _ := segment.Data.GetFloat64("longitude")
				location := tgbotapi.NewLocation(chatId, latitude, longitude)
				location.ReplyToMessageID = replyId
				results = append(results, location)
			}
		case libonebot.SegTypeReply:
			msgID, err := segment.Data.GetString("message_id")
			if err != nil {
				log.Errorln("msgid不存在，将忽略消息段 " + segment.Type)
				continue
			}
			id, err := strconv.Atoi(strings.Split(strings.Split(msgID, "&")[0], "_")[1])
			if err != nil {
				log.Errorln("msg_id错误" + msgID)
			}
			replyId = id

		case libonebot.SegTypeMention:
			{
				userIDStr, err := segment.Data.GetString("user_id")
				if err != nil {
					log.Errorln("msgid不存在，将忽略消息段 " + segment.Type)
					continue
				}
				userID, _ := strconv.ParseInt(userIDStr, 10, 64)
				member, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
					ChatID:             chatId,
					SuperGroupUsername: "",
					UserID:             userID,
				}})
				if err != nil {
					log.Errorln("member不存在，忽略消息段" + userIDStr)
					continue
				}
				if member.User.UserName == "" {
					member.User.UserName = member.User.FirstName + " " + member.User.LastName
				}
				mention += "[" + member.User.UserName + "](tg://user?id=" + strconv.FormatInt(member.User.ID, 10) + ")  "

			}
		}
	}
	var newResult []tgbotapi.Chattable
	for _, result := range results {
		if msg, ok := result.(tgbotapi.MessageConfig); ok {
			if mention != "" {
				message := tgbotapi.NewMessage(chatId, mention+msg.Text)
				message.ParseMode = tgbotapi.ModeMarkdownV2
				newResult = append(newResult, message)
			} else {
				newResult = append(newResult, result)
			}

		} else {
			newResult = append(newResult, result)
		}
	}
	return newResult
}
