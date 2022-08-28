package lib

import "github.com/botuniverse/go-libonebot"

type CustomComm func(ob *libonebot.OneBot, event libonebot.AnyEvent)

var (
	comms = make(map[string]CustomComm)
)

func RegisterCustomComm(name string, comm CustomComm) {
	comms[name] = comm
}
