package goproxy

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

const (
	upstreamProxy = "https://goproxy.cn"
)

func updateEnv(cmd *exec.Cmd) {
	cmd.Env = append(
		os.Environ(),
		"GOSUMDB=off",
	)
}

type ModVersionRsp struct {
	Version string    `json:"Version"`
	Time    time.Time `json:"Time"`
}

type Goproxy struct {
	client *resty.Client
	WebApp *fiber.App
}

func NewGoproxy() *Goproxy {
	return &Goproxy{}
}

func (p *Goproxy) Init() error {
	p.client = resty.New().
		SetTimeout(time.Second * 5).
		SetRetryMaxWaitTime(time.Second * 5).
		SetRetryWaitTime(time.Second).
		SetLogger(log.New()).
		SetDebug(true)

	p.WebApp = fiber.New(fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return ctx.Status(500).SendString(err.Error())
		},
		ServerHeader:  "",
		CaseSensitive: true,
		UnescapePath:  true,
		// ETag:                     true,
		ReadTimeout:              time.Minute * 5,
		WriteTimeout:             time.Minute * 5,
		CompressedFileSuffix:     ".gz",
		ProxyHeader:              "",
		DisableDefaultDate:       true,
		DisableHeaderNormalizing: true,
		ReduceMemoryUsage:        true,
	})

	p.WebApp.Server().Logger = log.New()

	p.WebApp.Get("*/@latest", func(ctx *fiber.Ctx) error {
		modInfo, err := GetModBaseInfoFromLocal(strings.TrimPrefix(strings.TrimSuffix(ctx.Path(), "/@latest")+"@latest", "/"))
		if err != nil {
			log.Errorf("err:%v", err)

			resp, err := p.client.R().Get(fmt.Sprintf("%s%s", upstreamProxy, ctx.Path()))
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
			return ctx.SendString(resp.String())
		}

		return ctx.JSON(&ModVersionRsp{
			Version: modInfo.Version,
			Time:    modInfo.Time,
		})
	})

	p.WebApp.Get("*/@v/list", func(ctx *fiber.Ctx) error {
		c := context.TODO()
		c, _ = context.WithTimeout(c, time.Second*5)

		goModVersions, err := GetModVersionsFromLocal(strings.TrimPrefix(strings.TrimSuffix(ctx.Path(), "/@v/list"), "/"))
		if err != nil {
			log.Errorf("err:%v", err)

			resp, err := p.client.R().Get(fmt.Sprintf("%s%s", upstreamProxy, ctx.Path()))
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
			return ctx.SendString(resp.String())
		}

		return ctx.SendString(goModVersions.Versions.Join("\n"))
	})

	p.WebApp.Get("*/@v/:version.info", func(ctx *fiber.Ctx) error {
		modInfo, err := GetModBaseInfoFromLocal(strings.TrimPrefix(strings.TrimSuffix(ctx.Path(), fmt.Sprintf("/@v/%s.info", ctx.Params("version"))), "/") + "@" + ctx.Params("version"))
		if err != nil {
			log.Errorf("err:%v", err)

			resp, err := p.client.R().Get(fmt.Sprintf("%s%s", upstreamProxy, ctx.Path()))
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
			return ctx.SendString(resp.String())
		}

		return ctx.JSON(&ModVersionRsp{
			Version: modInfo.Version,
			Time:    modInfo.Time,
		})
	})

	p.WebApp.Get("*/@v/:version.mod", func(ctx *fiber.Ctx) error {
		modInfo, err := GetModInfoFromLocal(strings.TrimPrefix(strings.TrimSuffix(ctx.Path(), fmt.Sprintf("/@v/%s.mod", ctx.Params("version"))), "/") + "@" + ctx.Params("version"))
		if err != nil {
			log.Errorf("err:%v", err)

			resp, err := p.client.R().Get(fmt.Sprintf("%s%s", upstreamProxy, ctx.Path()))
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
			return ctx.SendString(resp.String())
		}

		goMod, err := ioutil.ReadFile(modInfo.GoMod)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		ctx.Response().Header.SetContentType(fiber.MIMETextPlainCharsetUTF8)
		return ctx.SendString(string(goMod))
	})

	p.WebApp.Get("*/@v/:version.zip", func(ctx *fiber.Ctx) error {
		modInfo, err := GetModInfoFromLocal(strings.TrimPrefix(strings.TrimSuffix(ctx.Path(), fmt.Sprintf("/@v/%s.zip", ctx.Params("version"))), "/") + "@" + ctx.Params("version"))
		if err != nil {
			log.Errorf("err:%v", err)

			resp, err := p.client.R().Get(fmt.Sprintf("%s%s", upstreamProxy, ctx.Path()))
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
			return ctx.SendString(resp.String())
		}

		ctx.Response().Header.SetContentType("application/zip")
		zipFile, err := ioutil.ReadFile(modInfo.Zip)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		ctx.Response().Header.SetContentType(fiber.MIMETextPlainCharsetUTF8)
		return ctx.SendString(string(zipFile))
	})

	return nil
}

func (p *Goproxy) Run() error {
	return p.WebApp.Listen("0.0.0.0:19704")
}

func Start() error {
	p := NewGoproxy()

	err := p.Init()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = p.Run()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}
