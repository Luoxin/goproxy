package main

import (
	"os"

	"github.com/Luoxin/goproxy/goproxy"
	log "github.com/sirupsen/logrus"
)

func main() {
	os.Setenv("GOPROXY", "https://goproxy.cn,https://admin:pinkuai1228@goproxy.aquanliang.com,direct")

	err := goproxy.Start()
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}
}
