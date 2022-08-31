package lib

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/botuniverse/go-libonebot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"

	"github.com/huoxue1/tgonebot/utils"
)

const (
	// ActionSetGroupBan 群禁言
	ActionSetGroupBan = "set_group_ban"

	// ActionSetGroupKick 群踢人
	ActionSetGroupKick = "set_group_kick"

	// ActionGetCommands 获取命令列表
	ActionGetCommands = "get_commands"

	// ActionSetCommands 设置命令列表
	ActionSetCommands = "set_commands"

	// ActionEditTextMessage 编辑消息
	ActionEditTextMessage = "edit_text_message"

	// ActionSetInlineKeyBoard 设置交互按钮
	ActionSetInlineKeyBoard = "set_inline_key_board"
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
		detailType, err := req.Params.GetString("detail_type")
		if err != nil {
			log.Errorln("[send_message] detail_type")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		var idFrom string
		switch detailType {
		case "private":
			idFrom = "user_id"
		case "group":
			idFrom = "group_id"
		case "channel ":
			idFrom = "channel_id"
		}
		userIDStr, err := req.Params.GetString(idFrom)
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
		chattables := utils.MessageToChattables(bot, msgs, userID)
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
		//msgIDMap, err := req.Params.GetMap("message_id")
		//if err != nil {
		//	log.Errorln("[delete_message] 获取messageID失败")
		//	res.WriteFailed(libonebot.RetCodeBadParam, err)
		//	return
		//}
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
			_, err := bot.Request(tgbotapi.NewDeleteMessage(chatID, int(messageID)))
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
	// get_group_info
	mux.HandleFunc(libonebot.ActionGetGroupInfo, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		groupIdStr, err := req.Params.GetString("group_id")
		if err != nil {
			log.Errorln("[get_group_info] " + "group_id不存在")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		id, _ := strconv.ParseInt(groupIdStr, 10, 64)
		chat, err := bot.GetChat(tgbotapi.ChatInfoConfig{ChatConfig: tgbotapi.ChatConfig{ChatID: id}})
		if err != nil {
			res.WriteFailed(libonebot.RetCodeExecutionErrorBase, err)
			return
		}
		res.WriteData(map[string]any{
			"group_id":   groupIdStr,
			"group_name": chat.UserName,
		})
	})

	// get_group_member_info
	mux.HandleFunc(libonebot.ActionGetGroupMemberInfo, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		groupIdStr, err := req.Params.GetString("group_id")
		if err != nil {
			log.Errorln("[get_group_member_info] " + "group_id不存在")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		id, _ := strconv.ParseInt(groupIdStr, 10, 64)
		userIdStr, err := req.Params.GetString("user_id")
		if err != nil {
			log.Errorln("[get_group_member_info] " + "user_id不存在")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		userId, _ := strconv.ParseInt(userIdStr, 10, 64)
		member, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID:             id,
			SuperGroupUsername: "",
			UserID:             userId,
		}})
		if err != nil {
			log.Errorln("[get_group_member_info] 执行操作失败 " + err.Error())
			res.WriteFailed(libonebot.RetCodeExecutionErrorBase, err)
			return
		}
		res.WriteData(map[string]any{
			"user_id":          member.User.ID,
			"user_name":        member.User.UserName,
			"user_displayname": "",
		})
	})

	// get_file
	mux.HandleFunc(libonebot.ActionGetFile, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		fileID, err := req.Params.GetString("file_id")
		if err != nil {
			log.Errorln("[get_file] 获取fileID参数错误" + err.Error())
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		getType, err := req.Params.GetString("type")
		if err != nil {
			log.Errorln("[get_file] 获取type参数错误" + err.Error())
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		if getType != "url" {
			res.WriteFailed(libonebot.RetCodeUnsupportedParam, errors.New("un support param "+getType))
			return
		}
		url, err := bot.GetFileDirectURL(fileID)
		if err != nil {
			log.Errorln("[get_file] 获取文件错误" + err.Error())
			res.WriteFailed(libonebot.RetCodeExecutionErrorBase, err)
			return
		}
		res.WriteData(map[string]any{
			"name": "",
			"url":  url,
		})

	})
	// set_group_ban
	/*
	 * group_id : 群号 string
	 * user_id : 用户id string
	 * duration: 禁言时长。单位秒 int64
	 */
	mux.HandleFunc(ActionSetGroupBan, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		groupIdStr, err := req.Params.GetString("group_id")
		if err != nil {
			log.Errorln("[set_group_ban] " + "group_id不存在")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		id, _ := strconv.ParseInt(groupIdStr, 10, 64)
		userIdStr, err := req.Params.GetString("user_id")
		if err != nil {
			log.Errorln("[set_group_ban] " + "user_id不存在")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		duration, err := req.Params.GetInt64("duration")
		if err != nil {
			log.Errorln("[set_group_ban] " + "duration不存在")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		userId, _ := strconv.ParseInt(userIdStr, 10, 64)
		var config tgbotapi.Chattable
		if duration > 0 {
			config = tgbotapi.RestrictChatMemberConfig{
				ChatMemberConfig: tgbotapi.ChatMemberConfig{
					ChatID: id,
					UserID: userId,
				},
				UntilDate: time.Now().Unix() + duration,
				Permissions: &tgbotapi.ChatPermissions{
					CanSendMessages:       false,
					CanSendMediaMessages:  false,
					CanSendPolls:          false,
					CanSendOtherMessages:  false,
					CanAddWebPagePreviews: false,
					CanChangeInfo:         false,
					CanInviteUsers:        false,
					CanPinMessages:        false,
				},
			}
		} else {
			config = tgbotapi.RestrictChatMemberConfig{
				ChatMemberConfig: tgbotapi.ChatMemberConfig{
					ChatID: id,
					UserID: userId,
				},
				UntilDate: 9999999999999,
				Permissions: &tgbotapi.ChatPermissions{
					CanSendMessages:       true,
					CanSendMediaMessages:  true,
					CanSendPolls:          true,
					CanSendOtherMessages:  true,
					CanAddWebPagePreviews: true,
					CanChangeInfo:         true,
					CanInviteUsers:        true,
					CanPinMessages:        true,
				},
			}
		}
		_, err = bot.Request(config)
		if err != nil {
			log.Errorln("[set_group_ban] 执行失败" + err.Error())
			res.WriteFailed(libonebot.RetCodeExecutionErrorBase, err)
			return
		}
		res.WriteOK()

	})
	// set_group_kick
	/*
	 * group_id : 群号 string
	 * user_id : 用户id string
	 */
	mux.HandleFunc(ActionSetGroupKick, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		groupIdStr, err := req.Params.GetString("group_id")
		if err != nil {
			log.Errorln("[set_group_kick] " + "group_id不存在")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		id, _ := strconv.ParseInt(groupIdStr, 10, 64)
		userIdStr, err := req.Params.GetString("user_id")
		if err != nil {
			log.Errorln("[set_group_kick] " + "user_id不存在")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		userId, _ := strconv.ParseInt(userIdStr, 10, 64)
		_, err = bot.Request(tgbotapi.BanChatMemberConfig{
			ChatMemberConfig: tgbotapi.ChatMemberConfig{ChatID: id, UserID: userId},
			UntilDate:        0,
			RevokeMessages:   false,
		})
		if err != nil {
			log.Errorln("[set_group_kick] 执行失败" + err.Error())
			res.WriteFailed(libonebot.RetCodeExecutionErrorBase, err)
			return
		}
	})
	// get_commands
	/*
	 * result []BotCommand  command description
	 *
	 */
	mux.HandleFunc(ActionGetCommands, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		commands, err := bot.GetMyCommands()
		if err != nil {
			log.Errorln("[get_commands] 执行失败" + err.Error())
			res.WriteFailed(libonebot.RetCodeExecutionErrorBase, err)
			return
		}
		res.WriteData(commands)
	})

	// set_commands
	/*
	 * set_commands []BotCommand
	 *
	 */
	mux.HandleFunc(ActionSetCommands, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		array, err := req.Params.GetArray("commands")
		if err != nil {
			log.Errorln("[set_commands] " + "commands不存在")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		var commands []tgbotapi.BotCommand
		for _, easier := range array {
			easierMap := easier.(map[string]any)
			command, ok := easierMap["command"].(string)
			if !ok {
				log.Errorln("[set_commands] " + "command不存在，忽略该字段")
				continue
			}
			description, ok := easierMap["description"].(string)
			if err != nil {
				log.Errorln("[set_commands] " + "description不存在，忽略该字段")
				continue
			}
			commands = append(commands, tgbotapi.BotCommand{
				Command:     command,
				Description: description,
			})
		}
		_, err = bot.Request(tgbotapi.NewSetMyCommands(commands...))
		if err != nil {
			log.Errorln("[set_commands] 执行失败" + err.Error())
			res.WriteFailed(libonebot.RetCodeExecutionErrorBase, err)
			return
		}
	})
	// edit_text_message 编辑文本消息
	/*
	 * message_id : 消息id string
	 * text : 文本内容 string
	 */
	mux.HandleFunc(ActionEditTextMessage, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		msgID, err := req.Params.GetString("message_id")
		if err != nil {
			log.Errorln("[edit_text_message] 获取messageID失败")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		text, err := req.Params.GetString("text")
		if err != nil {
			log.Errorln("[edit_text_message] 获取text失败")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		msgIDs := strings.Split(msgID, "&")

		ids := strings.Split(msgIDs[0], "_")
		chatID, _ := strconv.ParseInt(ids[0], 10, 64)
		messageID, _ := strconv.ParseInt(ids[1], 10, 64)
		_, err = bot.Request(tgbotapi.NewEditMessageText(chatID, int(messageID), text))
		tgbotapi.NewInlineKeyboardMarkup()
		if err != nil {
			log.Errorln("[edit_text_message] 撤回消息错误" + err.Error())
			res.WriteFailed(libonebot.RetCodeExecutionErrorBase, err)
		}
		res.WriteOK()
	})
	// set_inline_key_board" 编辑文本消息
	/*
	 * message_id : 消息id string
	 * key_board [][]tgbotapi.InlineKeyboardButton  text,url or data
	 */
	mux.HandleFunc(ActionSetInlineKeyBoard, func(res libonebot.ResponseWriter, req *libonebot.Request) {
		msgID, err := req.Params.GetString("message_id")
		if err != nil {
			log.Errorln("[edit_text_message] 获取messageID失败")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}

		msgIDs := strings.Split(msgID, "&")

		ids := strings.Split(msgIDs[0], "_")
		chatID, _ := strconv.ParseInt(ids[0], 10, 64)
		messageID, _ := strconv.ParseInt(ids[1], 10, 64)

		array, err := req.Params.GetArray("key_board")
		if err != nil {
			log.Errorln("[edit_text_message] 获取key_board失败")
			res.WriteFailed(libonebot.RetCodeBadParam, err)
			return
		}
		var buttons [][]tgbotapi.InlineKeyboardButton
		for _, key := range array {
			row := key.([]interface{})
			var rows []tgbotapi.InlineKeyboardButton
			for _, r := range row {
				m := r.(map[string]any)
				var text, url, data string
				if t, ok := m["text"]; ok {
					text = t.(string)
				} else {
					log.Errorln("[edit_text_message] 确实text参数，忽略该字段")
					continue
				}
				if u, ok := m["url"]; ok {
					url = u.(string)
				}
				if d, ok := m["data"]; ok {
					data = d.(string)
				}

				button := tgbotapi.InlineKeyboardButton{
					Text:         text,
					URL:          &url,
					CallbackData: &data,
				}
				rows = append(rows, button)
			}
			buttons = append(buttons, rows)
		}
		_, err = bot.Request(tgbotapi.NewEditMessageReplyMarkup(chatID, int(messageID), tgbotapi.NewInlineKeyboardMarkup(buttons...)))
		if err != nil {
			log.Errorln("[edit_text_message] 执行失败")
			res.WriteFailed(libonebot.RetCodeExecutionErrorBase, err)
			return
		}
		res.WriteOK()
	})

	ob.Handle(mux)
}
