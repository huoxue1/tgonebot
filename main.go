package main

import (
	"encoding/json"

	"github.com/botuniverse/go-libonebot"
	log "github.com/sirupsen/logrus"

	"github.com/huoxue1/tgonebot/lib"
)

func main() {

	lib.RegisterCustomComm("custom", func(ob *libonebot.OneBot, event libonebot.AnyEvent) {
		if event.Name() == "message.group" {
			data, _ := json.Marshal(event)
			log.Info(string(data))
		}
	})

	lib.Main()
}
