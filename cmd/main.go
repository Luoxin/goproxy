package main

import (
	"github.com/Luoxin/goproxy/goproxy"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := goproxy.Start()
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}
}
