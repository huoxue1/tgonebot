package lib

import (
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/botuniverse/go-libonebot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	rotates "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"

	"github.com/huoxue1/tgonebot/conf"
)

const (
	Impl     = "tgonebot"
	Platform = "telegram"
	Version  = "0.0.1"
)

type TgOneBot struct {
	*libonebot.OneBot
}

func init() {
	conf.InitConfig("./config/config.yml")
}

func init() {
	logFormatter := &easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[%time%] [%lvl%]: %msg% \n",
	}
	w, err := rotates.New(path.Join("config", "logs", "%Y-%m-%d.log"), rotates.WithRotationTime(time.Hour*24))
	if err != nil {
		log.Errorf("rotates init err: %v", err)
		panic(err)
	}
	log.SetOutput(io.MultiWriter(w, os.Stdout))
	log.SetFormatter(logFormatter)
}

func Main() {
	config := conf.GetConfig()
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}
	bot, err := tgbotapi.NewBotAPIWithClient(config.Token, config.EndPoint+"/bot%s/%s", client)
	if err != nil {
		log.Errorln("初始化Bot出现异常" + err.Error())
		return
	}
	ob := libonebot.NewOneBot(Impl, Platform, strconv.FormatInt(bot.Self.ID, 10), &libonebot.Config{
		Heartbeat: config.Heartbeat,
		Comm:      config.Comm,
	})
	ob.Logger = log.StandardLogger()
	registerAction(bot, ob)
	handleEvent(ob, bot)
	ob.Run()
}
