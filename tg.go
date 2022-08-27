package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/botuniverse/go-libonebot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"

	"tgonebot/utils"
)

func registerAction(bot *tgbotapi.BotAPI, ob *libonebot.OneBot) {
	mux := libonebot.NewActionMux()
	mux.HandleFunc(libonebot.ActionGetVersion, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		res.WriteData(map[string]string{
			"impl":           Impl,
			"platform":       Platform,
			"version":        Version,
			"onebot_version": libonebot.OneBotVersion,
		})
	})

	// 注册 get_status 动作处理函数
	mux.HandleFunc(libonebot.ActionGetStatus, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		res.WriteData(map[string]interface{}{
			"good":   true,
			"online": true,
		})
	})
	// 注册 get_self_id 动作处理函数
	mux.HandleFunc(libonebot.ActionGetSelfInfo, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		res.WriteData(map[string]interface{}{
			"user_id":  bot.Self.ID,
			"nickname": bot.Self.UserName,
		})
	})
	// 发送消息
	mux.HandleFunc(libonebot.ActionSendMessage, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		msgs, err := req.Params.GetMessage("message")
		if err != nil {
			log.Errorln("[send_message] 获取消息段失败")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		userIDStr, err := req.Params.GetString("user_id")
		if err != nil {

			log.Errorln("[send_message] 获取userID失败")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return

		}
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			log.Errorln("[send_message] 转换userID错误")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
		}
		var msgIDs []string
		chattables := utils.MessageToChattables(msgs, userID)
		for _, chattable := range chattables {
			message, err := bot.Send(chattable)
			if err != nil {
				log.Errorln("发生消息错误" + err.Error())
				continue
			}
			msgIDs = append(msgIDs, userIDStr+"_"+strconv.Itoa(message.MessageID))
		}
		res.WriteData(map[string]interface{}{
			"message_id": strings.Join(msgIDs, "&"),
			"time":       time.Now(),
		})
	})
	// 撤回消息
	mux.HandleFunc(libonebot.ActionDeleteMessage, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		msgID, err := req.Params.GetString("message_id")
		if err != nil {
			log.Errorln("[delete_message] 获取messageID失败")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		var errs error
		msgIDs := strings.Split(msgID, "&")
		for _, id := range msgIDs {
			ids := strings.Split(id, "_")
			chatID, _ := strconv.ParseInt(ids[0], 10, 64)
			messageID, _ := strconv.ParseInt(ids[1], 10, 64)
			_, err := bot.Send(tgbotapi.NewDeleteMessage(chatID, int(messageID)))
			if err != nil {
				log.Errorln("[delete_message] 撤回消息错误" + err.Error())
				errs = err
				continue
			}
		}
		if errs != nil {
			res.WriteFailed(libonebot.RetCodeExecutionErrorBase, errs)
		}
		res.WriteOK()

	})

	mux.HandleFunc(libonebot.ActionGetSelfInfo, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		res.WriteData(map[string]interface{}{
			"user_id":          strconv.FormatInt(bot.Self.ID, 10),
			"user_name":        bot.Self.UserName,
			"user_displayname": "",
		})
	})
	mux.HandleFunc(libonebot.ActionGetGroupInfo, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		groupIdStr, err := req.Params.GetString("group_id")
		if err != nil {
			log.Errorln("[get_group_info] " + "group_id不存在")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		id, _ := strconv.ParseInt(groupIdStr, 10, 64)
		chat, err := bot.GetChat(tgbotapi.ChatInfoConfig{tgbotapi.ChatConfig{ChatID: id}})
		if err != nil {
			res.WriteFailed(libonebot.RetCodeExecutionErrorBase, err)
			return
		}
		res.WriteData(map[string]any{
			"group_id":   groupIdStr,
			"group_name": chat.UserName,
		})
	})
	ob.Handle(mux)
}
