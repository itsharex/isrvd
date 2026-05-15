package cron

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/goccy/go-yaml"
	"github.com/rehiy/libgo/logman"

	"isrvd/config"
)

var storeMu sync.Mutex

// loadJobs 从 cron.yml 加载任务列表
func loadJobs() ([]*Job, error) {
	data, err := os.ReadFile(cronFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var jobs []*Job
	if err := yaml.Unmarshal(data, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

// saveJobs 将任务列表写入 cron.yml
func saveJobs(jobs []*Job) error {
	storeMu.Lock()
	defer storeMu.Unlock()

	data, err := yaml.Marshal(jobs)
	if err != nil {
		return err
	}

	path := cronFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	logman.Debug("Save cron jobs", "path", path, "count", len(jobs))
	return os.WriteFile(path, data, 0644)
}

// cronFilePath 返回计划任务配置文件路径，存放在 server.rootDirectory 下。
func cronFilePath() string {
	return filepath.Join(config.Server.RootDirectory, "cron.yml")
}
