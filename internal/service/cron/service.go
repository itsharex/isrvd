// Package cron 计划任务业务服务
package cron

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/rehiy/libgo/command"
	"github.com/rehiy/libgo/logman"
	"github.com/rehiy/libgo/signal"
	cronlib "github.com/robfig/cron/v3"
)

// TypeInfo 脚本类型描述
type TypeInfo struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Job 计划任务业务类型
type Job struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Schedule    string `yaml:"schedule" json:"schedule"`
	Type        string `yaml:"type" json:"type"` // SHELL | EXEC | BAT | POWERSHELL
	Content     string `yaml:"content" json:"content"`
	WorkDir     string `yaml:"workDir" json:"workDir"`
	Timeout     uint   `yaml:"timeout" json:"timeout"` // 秒，0 表示不限制
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	Description string `yaml:"description" json:"description"`
}

// JobDetail 任务详情（含运行时调度状态）
type JobDetail struct {
	*Job
	NextRun *time.Time `json:"nextRun,omitempty"`
	LastRun *time.Time `json:"lastRun,omitempty"`
}

// JobLog 任务执行日志
type JobLog struct {
	JobID     string    `json:"jobId"`
	JobName   string    `json:"jobName"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Duration  int64     `json:"duration"` // 毫秒
	Success   bool      `json:"success"`
	Output    string    `json:"output"`
	Error     string    `json:"error,omitempty"`
}

// Service 计划任务服务
type Service struct {
	mu      sync.RWMutex
	cron    *cronlib.Cron
	jobs    map[string]*Job            // jobID → Job
	entries map[string]cronlib.EntryID // jobID → cron entry ID
	logs    []*JobLog                  // 执行历史（内存，最多保留 maxLogs 条）
	maxLogs int
}

// AvailableTypes 按当前 OS 返回可用脚本类型
func AvailableTypes() []TypeInfo {
	if runtime.GOOS == "windows" {
		return []TypeInfo{
			{Value: "BAT", Label: "BAT（批处理脚本）"},
			{Value: "POWERSHELL", Label: "POWERSHELL（PowerShell 脚本）"},
			{Value: "EXEC", Label: "EXEC（直接执行命令）"},
		}
	}
	return []TypeInfo{
		{Value: "SHELL", Label: "SHELL（Shell 脚本）"},
		{Value: "EXEC", Label: "EXEC（直接执行命令）"},
	}
}

// NewService 创建计划任务服务并启动调度器
func NewService() *Service {
	s := &Service{
		jobs:    make(map[string]*Job),
		entries: make(map[string]cronlib.EntryID),
		maxLogs: 500,
		cron:    cronlib.New(),
	}

	// 从 cron.yml 加载任务
	jobs, err := loadJobs()
	if err != nil {
		logman.Warn("Load cron jobs failed", "error", err)
	}
	for _, job := range jobs {
		if err := validateJob(job); err != nil {
			logman.Warn("Skip invalid cron job", "error", err)
			continue
		}
		s.jobs[job.ID] = job
		if job.Enabled {
			if err := s.register(job); err != nil {
				logman.Warn("Cron job register failed", "id", job.ID, "name", job.Name, "error", err)
			}
		}
	}

	s.cron.Start()
	logman.Info("Cron scheduler started", "jobs", len(s.entries))

	signal.OnQuit(func() {
		ctx := s.cron.Stop()
		<-ctx.Done()
		logman.Info("Cron scheduler stopped")
	})

	return s
}

// ─── 公开方法 ───

// ListJobs 返回所有任务（含运行状态）
func (s *Service) ListJobs() []*JobDetail {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*JobDetail, 0, len(s.jobs))
	for _, job := range s.jobs {
		detail := &JobDetail{Job: job}
		if entryID, ok := s.entries[job.ID]; ok {
			e := s.cron.Entry(entryID)
			if !e.Next.IsZero() {
				detail.NextRun = &e.Next
			}
			if !e.Prev.IsZero() {
				detail.LastRun = &e.Prev
			}
		}
		result = append(result, detail)
	}
	return result
}

// CreateJob 创建任务并持久化
func (s *Service) CreateJob(job *Job) error {
	if err := validateJob(job); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[job.ID]; exists {
		return fmt.Errorf("job already exists: %s", job.ID)
	}

	s.jobs[job.ID] = job
	if job.Enabled {
		if err := s.register(job); err != nil {
			delete(s.jobs, job.ID)
			return err
		}
	}

	if err := s.persist(); err != nil {
		if entryID, ok := s.entries[job.ID]; ok {
			s.cron.Remove(entryID)
			delete(s.entries, job.ID)
		}
		delete(s.jobs, job.ID)
		return err
	}
	return nil
}

// UpdateJob 更新任务并重新注册
func (s *Service) UpdateJob(job *Job) error {
	if err := validateJob(job); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldJob, ok := s.jobs[job.ID]
	if !ok {
		return fmt.Errorf("job not found: %s", job.ID)
	}
	oldEntryID, oldEnabled := s.entries[job.ID]

	if oldEnabled {
		s.cron.Remove(oldEntryID)
		delete(s.entries, job.ID)
	}
	s.jobs[job.ID] = job

	if job.Enabled {
		if err := s.register(job); err != nil {
			s.jobs[job.ID] = oldJob
			if oldEnabled {
				if oldEntryID, err := s.cron.AddFunc(oldJob.Schedule, func() { s.runJob(oldJob.ID) }); err == nil {
					s.entries[oldJob.ID] = oldEntryID
				}
			}
			return err
		}
	}

	if err := s.persist(); err != nil {
		if entryID, ok := s.entries[job.ID]; ok {
			s.cron.Remove(entryID)
			delete(s.entries, job.ID)
		}
		s.jobs[job.ID] = oldJob
		if oldEnabled {
			if entryID, err := s.cron.AddFunc(oldJob.Schedule, func() { s.runJob(oldJob.ID) }); err == nil {
				s.entries[oldJob.ID] = entryID
			}
		}
		return err
	}
	return nil
}

// DeleteJob 删除任务
func (s *Service) DeleteJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[id]
	if !ok {
		return fmt.Errorf("job not found: %s", id)
	}

	oldEntryID, oldEnabled := s.entries[id]
	if oldEnabled {
		s.cron.Remove(oldEntryID)
		delete(s.entries, id)
	}
	delete(s.jobs, id)

	if err := s.persist(); err != nil {
		s.jobs[id] = job
		if oldEnabled {
			if entryID, addErr := s.cron.AddFunc(job.Schedule, func() { s.runJob(job.ID) }); addErr == nil {
				s.entries[id] = entryID
			}
		}
		return err
	}
	return nil
}

// ToggleJob 启用或禁用任务
func (s *Service) ToggleJob(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[id]
	if !ok {
		return fmt.Errorf("job not found: %s", id)
	}

	if job.Enabled == enabled {
		return nil
	}

	oldEnabled := job.Enabled
	job.Enabled = enabled

	if enabled {
		if err := s.register(job); err != nil {
			job.Enabled = oldEnabled
			return err
		}
	} else {
		if entryID, ok := s.entries[id]; ok {
			s.cron.Remove(entryID)
			delete(s.entries, id)
		}
	}

	if err := s.persist(); err != nil {
		job.Enabled = oldEnabled
		if enabled {
			if entryID, ok := s.entries[id]; ok {
				s.cron.Remove(entryID)
				delete(s.entries, id)
			}
		} else if oldEnabled {
			if entryID, err := s.cron.AddFunc(job.Schedule, func() { s.runJob(job.ID) }); err == nil {
				s.entries[id] = entryID
			}
		}
		return err
	}
	return nil
}

// RunNow 立即触发一次任务（异步执行）
func (s *Service) RunNow(id string) error {
	s.mu.RLock()
	_, ok := s.jobs[id]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("job not found: %s", id)
	}

	go s.runJob(id)
	return nil
}

// GetLogs 返回指定任务的执行历史（最近 limit 条，倒序）
func (s *Service) GetLogs(id string, limit int) []*JobLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*JobLog
	for i := len(s.logs) - 1; i >= 0 && len(result) < limit; i-- {
		if id == "" || s.logs[i].JobID == id {
			result = append(result, s.logs[i])
		}
	}
	return result
}

// ─── 内部方法 ───

func validateJob(job *Job) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}
	if job.ID == "" {
		return fmt.Errorf("job id is required")
	}
	if job.Name == "" {
		return fmt.Errorf("job name is required")
	}
	if job.Schedule == "" {
		return fmt.Errorf("job schedule is required")
	}
	if _, err := cronlib.ParseStandard(job.Schedule); err != nil {
		return fmt.Errorf("invalid schedule %q: %w", job.Schedule, err)
	}
	typeAllowed := false
	for _, item := range AvailableTypes() {
		if item.Value == job.Type {
			typeAllowed = true
			break
		}
	}
	if !typeAllowed {
		return fmt.Errorf("unsupported script type on %s: %s", runtime.GOOS, job.Type)
	}
	if job.Content == "" {
		return fmt.Errorf("job content is required")
	}
	return nil
}

// register 向调度器注册一个任务（调用前须持有锁或在初始化阶段）
func (s *Service) register(job *Job) error {
	entryID, err := s.cron.AddFunc(job.Schedule, func() { s.runJob(job.ID) })
	if err != nil {
		return fmt.Errorf("invalid schedule %q: %w", job.Schedule, err)
	}
	s.entries[job.ID] = entryID
	return nil
}

// persist 将当前 jobs 持久化到 cron.yml（调用前须持有锁）
func (s *Service) persist() error {
	jobs := make([]*Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		jobs = append(jobs, j)
	}
	return saveJobs(jobs)
}

// runJob 执行指定 ID 的任务
func (s *Service) runJob(id string) {
	s.mu.RLock()
	job, ok := s.jobs[id]
	s.mu.RUnlock()
	if !ok {
		return
	}

	start := time.Now()
	logman.Info("Cron job running", "id", job.ID, "name", job.Name)

	output, err := command.RunScript(&command.ScriptPayload{
		Name:       job.Name,
		ScriptType: job.Type,
		Content:    job.Content,
		WorkDir:    job.WorkDir,
		Timeout:    job.Timeout,
	})
	end := time.Now()

	entry := &JobLog{
		JobID:     job.ID,
		JobName:   job.Name,
		StartTime: start,
		EndTime:   end,
		Duration:  end.Sub(start).Milliseconds(),
		Success:   err == nil,
		Output:    output,
	}
	if err != nil {
		entry.Error = err.Error()
		logman.Warn("Cron job failed", "id", job.ID, "name", job.Name, "error", err)
	} else {
		logman.Info("Cron job done", "id", job.ID, "name", job.Name, "duration", entry.Duration)
	}

	s.mu.Lock()
	s.logs = append(s.logs, entry)
	if len(s.logs) > s.maxLogs {
		s.logs = s.logs[len(s.logs)-s.maxLogs:]
	}
	s.mu.Unlock()
}
