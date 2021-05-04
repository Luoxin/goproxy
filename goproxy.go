package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/elliotchance/pie/pie"
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"

	"github.com/Luoxin/Eutamias/utils"
)

func init() {
	execPath := utils.GetExecPath()
	logPath := filepath.Join(execPath, "goproxy.log")

	writer, err := rotatelogs.New(
		filepath.Join(execPath, "eutamias-%Y-%m-%d-%H.log"),
		rotatelogs.WithLinkName(logPath),
		rotatelogs.WithMaxAge(time.Hour),
		rotatelogs.WithRotationTime(time.Minute),
	)
	if err != nil {
		log.Fatalf("err:%v", err)
	}

	log.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			log.InfoLevel:  writer,
			log.WarnLevel:  writer,
			log.ErrorLevel: writer,
			log.FatalLevel: writer,
			log.PanicLevel: writer,
		},
		&nested.Formatter{
			FieldsOrder: []string{
				log.FieldKeyTime, log.FieldKeyLevel, log.FieldKeyFile,
				log.FieldKeyFunc, log.FieldKeyMsg,
			},
			CustomCallerFormatter: func(f *runtime.Frame) string {
				return fmt.Sprintf("(%s %s:%d)", f.Function, path.Base(f.File), f.Line)
			},
			TimestampFormat:  time.RFC3339,
			HideKeys:         true,
			NoFieldsSpace:    true,
			NoUppercaseLevel: true,
			TrimMessages:     true,
			CallerFirst:      true,
		},
	))
	log.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			log.TraceLevel: os.Stdout,
			log.DebugLevel: os.Stdout,
			log.InfoLevel:  os.Stdout,
			log.WarnLevel:  os.Stdout,
			log.ErrorLevel: os.Stdout,
			log.FatalLevel: os.Stdout,
			log.PanicLevel: os.Stdout,
		},
		&nested.Formatter{
			FieldsOrder: []string{
				log.FieldKeyTime, log.FieldKeyLevel, log.FieldKeyFile,
				log.FieldKeyFunc, log.FieldKeyMsg,
			},
			CustomCallerFormatter: func(f *runtime.Frame) string {
				return fmt.Sprintf("(%s %s:%d)", f.Function, path.Base(f.File), f.Line)
			},
			TimestampFormat:  time.RFC3339,
			HideKeys:         true,
			NoFieldsSpace:    true,
			NoUppercaseLevel: true,
			TrimMessages:     true,
			CallerFirst:      true,
		},
	))

	log.SetFormatter(&nested.Formatter{
		FieldsOrder: []string{
			log.FieldKeyTime, log.FieldKeyLevel, log.FieldKeyFile,
			log.FieldKeyFunc, log.FieldKeyMsg,
		},
		CustomCallerFormatter: func(f *runtime.Frame) string {
			return fmt.Sprintf("(%s %s:%d)", f.Function, path.Base(f.File), f.Line)
		},
		TimestampFormat:  time.RFC3339,
		HideKeys:         true,
		NoFieldsSpace:    true,
		NoUppercaseLevel: true,
		TrimMessages:     true,
		CallerFirst:      true,
	})
	log.SetReportCaller(true)
}

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

func main() {
	client := resty.New().
		SetTimeout(time.Second * 5).
		SetRetryMaxWaitTime(time.Second * 5).
		SetRetryWaitTime(time.Second).
		SetLogger(log.New()).
		SetDebug(true)

	app := fiber.New(fiber.Config{
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

	app.Server().Logger = log.New()

	app.Get("*/@latest", func(ctx *fiber.Ctx) error {
		c := context.TODO()
		c, _ = context.WithTimeout(c, time.Second*5)

		cmd := exec.CommandContext(c, "go", "list", "-json", "-m", strings.TrimPrefix(strings.TrimSuffix(ctx.Path(), "/@latest")+"@latest", "/"))
		updateEnv(cmd)
		stdout, err := cmd.Output()
		if err != nil {
			log.Error(err)
			if err := c.Err(); errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("command %v: %w", cmd.Args, err)
			}

			output := stdout
			if len(output) > 0 {
				m := map[string]interface{}{}
				if err := json.Unmarshal(output, &m); err != nil {
					return err
				}

				if es, ok := m["Error"].(string); ok {
					output = []byte(es)
				}
			} else if ee, ok := err.(*exec.ExitError); ok {
				output = ee.Stderr
			} else {
				return err
			}

			log.Error(string(output))

			resp, err := client.R().Get(fmt.Sprintf("%s%s", upstreamProxy, ctx.Path()))
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
			return ctx.SendString(resp.String())
		}

		log.Infof("get:%v", string(stdout))

		var goModList struct {
			Path      string    `json:"Path"`
			Version   string    `json:"Version"`
			Time      time.Time `json:"Time"`
			GoMod     string    `json:"GoMod"`
			GoVersion string    `json:"GoVersion"`
		}
		err = json.Unmarshal(stdout, &goModList)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		return ctx.JSON(&ModVersionRsp{
			Version: goModList.Version,
			Time:    goModList.Time,
		})
	})

	app.Get("*/@v/list", func(ctx *fiber.Ctx) error {
		c := context.TODO()
		c, _ = context.WithTimeout(c, time.Second*5)

		cmd := exec.CommandContext(c, "go", "list", "-json", "-m", "-versions", strings.TrimPrefix(strings.TrimSuffix(ctx.Path(), "/@v/list"), "/"))
		updateEnv(cmd)
		stdout, err := cmd.Output()
		if err != nil {
			log.Error(err)
			if err := c.Err(); errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("command %v: %w", cmd.Args, err)
			}

			output := stdout
			if len(output) > 0 {
				m := map[string]interface{}{}
				if err := json.Unmarshal(output, &m); err != nil {
					return err
				}

				if es, ok := m["Error"].(string); ok {
					output = []byte(es)
				}
			} else if ee, ok := err.(*exec.ExitError); ok {
				output = ee.Stderr
			} else {
				return err
			}

			log.Error(string(output))

			resp, err := client.R().Get(fmt.Sprintf("%s%s", upstreamProxy, ctx.Path()))
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
			return ctx.SendString(resp.String())
		}

		log.Infof("get:%v", string(stdout))

		var goModVersions struct {
			Path      string      `json:"Path"`
			Version   string      `json:"Version"`
			Versions  pie.Strings `json:"Versions"`
			Time      time.Time   `json:"Time"`
			Dir       string      `json:"Dir"`
			GoMod     string      `json:"GoMod"`
			GoVersion string      `json:"GoVersion"`
		}
		err = json.Unmarshal(stdout, &goModVersions)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		return ctx.SendString(goModVersions.Versions.Join("\n"))
	})

	app.Get("*/@v/:version.info", func(ctx *fiber.Ctx) error {
		c := context.TODO()
		c, _ = context.WithTimeout(c, time.Second*5)

		// cmd := exec.CommandContext(c, "go", "mod", "download", "-json",
		// 	strings.TrimPrefix(strings.TrimSuffix(ctx.Path(), fmt.Sprintf("/@v/%s.info", ctx.Params("version"))), "/")+"@"+ctx.Params("version"))

		cmd := exec.CommandContext(c, "go", "list", "-json", "-m",
			strings.TrimPrefix(strings.TrimSuffix(ctx.Path(), fmt.Sprintf("/@v/%s.info", ctx.Params("version"))), "/")+"@"+ctx.Params("version"))

		updateEnv(cmd)
		stdout, err := cmd.Output()
		if err != nil {
			log.Error(err)
			if err := c.Err(); errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("command %v: %w", cmd.Args, err)
			}

			output := stdout
			if len(output) > 0 {
				m := map[string]interface{}{}
				if err := json.Unmarshal(output, &m); err != nil {
					return err
				}

				if es, ok := m["Error"].(string); ok {
					output = []byte(es)
				}
			} else if ee, ok := err.(*exec.ExitError); ok {
				output = ee.Stderr
			} else {
				return err
			}

			log.Error(string(output))

			resp, err := client.R().Get(fmt.Sprintf("%s%s", upstreamProxy, ctx.Path()))
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
			return ctx.SendString(resp.String())
		}

		log.Infof("get:%v", string(stdout))

		var goModInfo struct {
			Path     string    `json:"Path"`
			Version  string    `json:"Version"`
			Info     string    `json:"Info"`
			GoMod    string    `json:"GoMod"`
			Zip      string    `json:"Zip"`
			Dir      string    `json:"Dir"`
			Sum      string    `json:"Sum"`
			GoModSum string    `json:"GoModSum"`
			Time     time.Time `json:"Time"`
		}
		err = json.Unmarshal(stdout, &goModInfo)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		return ctx.JSON(&ModVersionRsp{
			Version: goModInfo.Version,
			Time:    goModInfo.Time,
		})
	})

	app.Get("*/@v/:version.mod", func(ctx *fiber.Ctx) error {
		c := context.TODO()
		c, _ = context.WithTimeout(c, time.Second*5)

		cmd := exec.CommandContext(c, "go", "mod", "download", "-json",
			strings.TrimPrefix(strings.TrimSuffix(ctx.Path(), fmt.Sprintf("/@v/%s.mod", ctx.Params("version"))), "/")+"@"+ctx.Params("version"))

		updateEnv(cmd)
		stdout, err := cmd.Output()
		if err != nil {
			log.Error(err)
			if err := c.Err(); errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("command %v: %w", cmd.Args, err)
			}

			output := stdout
			if len(output) > 0 {
				m := map[string]interface{}{}
				if err := json.Unmarshal(output, &m); err != nil {
					return err
				}

				if es, ok := m["Error"].(string); ok {
					output = []byte(es)
				}
			} else if ee, ok := err.(*exec.ExitError); ok {
				output = ee.Stderr
			} else {
				return err
			}

			log.Error(string(output))

			resp, err := client.R().Get(fmt.Sprintf("%s%s", upstreamProxy, ctx.Path()))
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
			return ctx.SendString(resp.String())
		}

		log.Infof("get:%v", string(stdout))

		var goModInfo struct {
			Path     string `json:"Path"`
			Version  string `json:"Version"`
			Info     string `json:"Info"`
			GoMod    string `json:"GoMod"`
			Zip      string `json:"Zip"`
			Dir      string `json:"Dir"`
			Sum      string `json:"Sum"`
			GoModSum string `json:"GoModSum"`
		}
		err = json.Unmarshal(stdout, &goModInfo)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		goMod, err := ioutil.ReadFile(goModInfo.GoMod)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		ctx.Response().Header.SetContentType(fiber.MIMETextPlainCharsetUTF8)
		return ctx.SendString(string(goMod))
	})

	app.Get("*/@v/:version.zip", func(ctx *fiber.Ctx) error {
		c := context.TODO()
		c, _ = context.WithTimeout(c, time.Second*5)

		cmd := exec.CommandContext(c, "go", "mod", "download", "-json",
			strings.TrimPrefix(strings.TrimSuffix(ctx.Path(), fmt.Sprintf("/@v/%s.mod", ctx.Params("version"))), "/")+"@"+ctx.Params("version"))

		updateEnv(cmd)
		stdout, err := cmd.Output()
		if err != nil {
			log.Error(err)
			if err := c.Err(); errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("command %v: %w", cmd.Args, err)
			}

			output := stdout
			if len(output) > 0 {
				m := map[string]interface{}{}
				if err := json.Unmarshal(output, &m); err != nil {
					return err
				}

				if es, ok := m["Error"].(string); ok {
					output = []byte(es)
				}
			} else if ee, ok := err.(*exec.ExitError); ok {
				output = ee.Stderr
			} else {
				return err
			}

			log.Error(string(output))

			resp, err := client.R().Get(fmt.Sprintf("%s%s", upstreamProxy, ctx.Path()))
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
			return ctx.SendString(resp.String())
		}

		log.Infof("get:%v", string(stdout))

		var goModInfo struct {
			Path     string `json:"Path"`
			Version  string `json:"Version"`
			Info     string `json:"Info"`
			GoMod    string `json:"GoMod"`
			Zip      string `json:"Zip"`
			Dir      string `json:"Dir"`
			Sum      string `json:"Sum"`
			GoModSum string `json:"GoModSum"`
		}
		err = json.Unmarshal(stdout, &goModInfo)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		ctx.Response().Header.SetContentType("application/zip")
		zipFile, err := ioutil.ReadFile(goModInfo.Zip)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		ctx.Response().Header.SetContentType(fiber.MIMETextPlainCharsetUTF8)
		return ctx.SendString(string(zipFile))
	})

	_ = app.Listen("0.0.0.0:19704")
}
