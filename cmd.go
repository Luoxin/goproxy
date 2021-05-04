package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

type ModInfo struct {
	Path      string    `json:"Path"`
	Version   string    `json:"Version"`
	Time      time.Time `json:"Time"`
	GoMod     string    `json:"GoMod"`
	GoVersion string    `json:"GoVersion"`
}

func GetModInfo(modPath string) (*ModInfo, error) {
	ctx := context.TODO()
	ctx, _ = context.WithTimeout(ctx, time.Second*5)

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

	var modIndo ModInfo
	err = json.Unmarshal(stdout, &modIndo)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return &modIndo, nil
}
