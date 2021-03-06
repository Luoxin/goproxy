package goproxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/elliotchance/pie/pie"
	log "github.com/sirupsen/logrus"
)

type ModBaseInfo struct {
	Path      string    `json:"Path"`
	Version   string    `json:"Version"`
	Time      time.Time `json:"Time"`
	GoMod     string    `json:"GoMod"`
	GoVersion string    `json:"GoVersion"`
}

type ModInfo struct {
	Path     string `json:"Path"`
	Version  string `json:"Version"`
	Info     string `json:"Info"`
	GoMod    string `json:"GoMod"`
	Zip      string `json:"Zip"`
	Dir      string `json:"Dir"`
	Sum      string `json:"Sum"`
	GoModSum string `json:"GoModSum"`
}

type ModVersions struct {
	Path      string      `json:"Path"`
	Version   string      `json:"Version"`
	Versions  pie.Strings `json:"Versions"`
	Time      time.Time   `json:"Time"`
	Dir       string      `json:"Dir"`
	GoMod     string      `json:"GoMod"`
	GoVersion string      `json:"GoVersion"`
}

func GetModBaseInfoFromLocal(modPath string) (*ModBaseInfo, error) {
	ctx := context.TODO()
	ctx, _ = context.WithTimeout(ctx, time.Minute)

	cmd := exec.CommandContext(ctx, "go", "list", "-json", "-m", modPath)
	updateEnv(cmd)
	stdout, err := cmd.Output()
	if err != nil {
		log.Error(err)
		if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("command %v: %w", cmd.Args, err)
		}

		output := stdout
		if len(output) > 0 {
			m := map[string]interface{}{}
			if err := json.Unmarshal(output, &m); err != nil {
				return nil, err
			}

			if es, ok := m["Error"].(string); ok {
				output = []byte(es)
			}
		} else if ee, ok := err.(*exec.ExitError); ok {
			output = ee.Stderr
		} else {
			return nil, err
		}

		return nil, errors.New(string(output))
	}

	log.Infof("get:%v", string(stdout))

	var modIndo ModBaseInfo
	err = json.Unmarshal(stdout, &modIndo)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return &modIndo, nil
}

func GetModInfoFromLocal(modPath string) (*ModInfo, error) {
	ctx := context.TODO()
	ctx, _ = context.WithTimeout(ctx, time.Minute)

	cmd := exec.CommandContext(ctx, "go", "mod", "download", "-json", modPath)
	updateEnv(cmd)
	stdout, err := cmd.Output()
	if err != nil {
		log.Error(err)
		if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("command %v: %w", cmd.Args, err)
		}

		output := stdout
		if len(output) > 0 {
			m := map[string]interface{}{}
			if err := json.Unmarshal(output, &m); err != nil {
				return nil, err
			}

			if es, ok := m["Error"].(string); ok {
				output = []byte(es)
			}
		} else if ee, ok := err.(*exec.ExitError); ok {
			output = ee.Stderr
		} else {
			return nil, err
		}

		return nil, errors.New(string(output))
	}

	log.Infof("get:%v", string(stdout))

	var modInfo ModInfo
	err = json.Unmarshal(stdout, &modInfo)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return &modInfo, nil
}

func GetModVersionsFromLocal(modPath string) (*ModVersions, error) {
	ctx := context.TODO()
	ctx, _ = context.WithTimeout(ctx, time.Minute)

	cmd := exec.CommandContext(ctx, "go", "list", "-json", "-m", "-versions", modPath)
	updateEnv(cmd)
	stdout, err := cmd.Output()
	if err != nil {
		log.Error(err)
		if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("command %v: %w", cmd.Args, err)
		}

		output := stdout
		if len(output) > 0 {
			m := map[string]interface{}{}
			if err := json.Unmarshal(output, &m); err != nil {
				return nil, err
			}

			if es, ok := m["Error"].(string); ok {
				output = []byte(es)
			}
		} else if ee, ok := err.(*exec.ExitError); ok {
			output = ee.Stderr
		} else {
			return nil, err
		}

		return nil, errors.New(string(output))
	}

	log.Infof("get:%v", string(stdout))

	var modVersions ModVersions
	err = json.Unmarshal(stdout, &modVersions)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return &modVersions, nil
}
