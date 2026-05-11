package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Paths struct {
	LogDir   string
	CacheDir string
	RunDir   string
}

type LayoutOptions struct {
	RootDir    string
	ConfigDir  string
	StateDir   string
	SiteConfig string
	Manifest   string
	LogDir     string
}

type Layout struct {
	RootDir    string
	ConfigDir  string
	StateDir   string
	LogDir     string
	CacheDir   string
	RunDir     string
	SiteConfig string
	Manifest   string
}

type Job struct {
	Name   string `mapstructure:"name"`
	Site   string `mapstructure:"site"`
	Action string `mapstructure:"action"`
	File   string `mapstructure:"file"`
}

type ExecuteBatchOptions struct {
	JobName string
}

type TaskRecord struct {
	ID            string `json:"id"`
	Action        string `json:"action"`
	Status        string `json:"status"`
	JobName       string `json:"job_name,omitempty"`
	SourceFile    string `json:"source_file,omitempty"`
	RunID         string `json:"run_id"`
	URLCount      int    `json:"url_count"`
	SubmittedAt   string `json:"submitted_at"`
	LastCheckedAt string `json:"last_checked_at,omitempty"`
	CompletedAt   string `json:"completed_at,omitempty"`
}

type TaskState struct {
	Site      string       `json:"site"`
	Action    string       `json:"action"`
	UpdatedAt string       `json:"updated_at"`
	Tasks     []TaskRecord `json:"tasks"`
}

type RunRecord struct {
	ID           string   `json:"id"`
	Site         string   `json:"site"`
	Action       string   `json:"action"`
	JobName      string   `json:"job_name,omitempty"`
	SourceFile   string   `json:"source_file,omitempty"`
	LogFile      string   `json:"log_file"`
	CacheFile    string   `json:"cache_file"`
	RunFile      string   `json:"run_file"`
	SubmittedIDs []string `json:"submitted_ids"`
	StartedAt    string   `json:"started_at"`
	FinishedAt   string   `json:"finished_at,omitempty"`
	Status       string   `json:"status"`
	Error        string   `json:"error,omitempty"`
}

func ResolveLayout(execPath string, opts LayoutOptions) (Layout, error) {
	execDir, err := executableDir(execPath)
	if err != nil {
		return Layout{}, err
	}

	rootDir := resolveFirstNonEmpty(opts.RootDir, os.Getenv("CDNFIX_ROOT"))
	resolvedRoot := ""
	if rootDir != "" {
		resolvedRoot = ResolvePath(execDir, rootDir)
	}
	baseDir := execDir
	if resolvedRoot != "" {
		baseDir = resolvedRoot
	}
	configDir := resolveDir(baseDir, resolvedRoot, opts.ConfigDir, "CDNFIX_CONFIG_DIR", filepath.Join("config"), "/etc/cdnfix")
	stateDir := resolveDir(baseDir, resolvedRoot, opts.StateDir, "CDNFIX_STATE_DIR", filepath.Join("var", "lib"), "/var/lib/cdnfix")
	logDir := resolveDir(baseDir, resolvedRoot, opts.LogDir, "CDNFIX_LOG_DIR", filepath.Join("var", "log"), "/var/log/cdnfix")
	cacheDir := filepath.Join(stateDir, "cache")
	runDir := filepath.Join(stateDir, "runs")
	siteConfig := resolveFirstExisting(
		ResolvePath(configDir, opts.SiteConfig),
		filepath.Join(configDir, "sites.yaml"),
		filepath.Join(configDir, ".env"),
	)
	manifest := resolveWithDefault(configDir, opts.Manifest, filepath.Join(configDir, "jobs.yaml"))

	layout := Layout{
		RootDir:    resolvedRoot,
		ConfigDir:  configDir,
		StateDir:   stateDir,
		LogDir:     logDir,
		CacheDir:   cacheDir,
		RunDir:     runDir,
		SiteConfig: siteConfig,
		Manifest:   manifest,
	}
	return layout, nil
}

func ReadJobs(path string) ([]Job, error) {
	cfg := viper.New()
	cfg.SetConfigFile(path)
	if ext := strings.TrimPrefix(filepath.Ext(path), "."); ext != "" {
		cfg.SetConfigType(ext)
	}
	if err := cfg.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read jobs config: %w", err)
	}
	var jobs []Job
	if err := cfg.UnmarshalKey("jobs", &jobs); err != nil {
		return nil, fmt.Errorf("failed to parse jobs: %w", err)
	}
	if len(jobs) == 0 {
		return nil, fmt.Errorf("jobs config is empty")
	}
	for i := range jobs {
		jobs[i].Name = strings.TrimSpace(jobs[i].Name)
		jobs[i].Site = strings.TrimSpace(jobs[i].Site)
		jobs[i].Action = normalizeAction(jobs[i].Action)
		jobs[i].File = ResolveJobFile(path, jobs[i].File)
		if jobs[i].Site == "" || jobs[i].Action == "" || jobs[i].File == "" {
			return nil, fmt.Errorf("invalid job at index %d", i)
		}
	}
	return jobs, nil
}

func ResolveJobFile(manifestPath string, jobFile string) string {
	return ResolvePath(filepath.Dir(manifestPath), strings.TrimSpace(jobFile))
}

func normalizeAction(action string) string {
	action = strings.ToLower(strings.TrimSpace(action))
	switch action {
	case "refresh", "push":
		return action
	default:
		return action
	}
}

func ValidateAction(action string) error {
	switch normalizeAction(action) {
	case "refresh", "push":
		return nil
	default:
		return fmt.Errorf("unsupported action %q", action)
	}
}

func NewRun(paths Paths, site string, action string, jobName string, sourceFile string, now time.Time) RunRecord {
	action = normalizeAction(action)
	namePart := sanitizeName(jobName)
	if strings.TrimSpace(jobName) == "" {
		namePart = "adhoc"
	}
	runID := fmt.Sprintf("%s.%s.%s.%d", sanitizeName(site), sanitizeName(action), namePart, now.UnixNano())
	dateDir := now.Format("2006-01-02")
	cacheFile := TaskStatePath(paths, site, action)
	logFile := filepath.Join(paths.LogDir, dateDir, runID+".log")
	runFile := filepath.Join(paths.RunDir, dateDir, runID+".json")
	return RunRecord{
		ID:         runID,
		Site:       site,
		Action:     action,
		JobName:    jobName,
		SourceFile: sourceFile,
		LogFile:    logFile,
		CacheFile:  cacheFile,
		RunFile:    runFile,
		StartedAt:  now.Format(time.RFC3339),
		Status:     "running",
	}
}

func TaskStatePath(paths Paths, site string, action string) string {
	return filepath.Join(paths.CacheDir, sanitizeName(site), fmt.Sprintf("%s.tasks.json", normalizeAction(action)))
}

func SaveRun(run RunRecord) error {
	run.FinishedAt = strings.TrimSpace(run.FinishedAt)
	return writeJSON(run.RunFile, run)
}

func LoadTaskState(path string, site string, action string) (*TaskState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &TaskState{Site: site, Action: normalizeAction(action), Tasks: []TaskRecord{}}, nil
		}
		return nil, fmt.Errorf("failed to read task state: %w", err)
	}
	var state TaskState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse task state: %w", err)
	}
	if state.Site == "" {
		state.Site = site
	}
	if state.Action == "" {
		state.Action = normalizeAction(action)
	}
	return &state, nil
}

func SaveTaskState(path string, state *TaskState) error {
	state.UpdatedAt = time.Now().Format(time.RFC3339)
	return writeJSON(path, state)
}

var stateMutex sync.Mutex

func AddTask(path string, site string, action string, task TaskRecord) error {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	state, err := LoadTaskState(path, site, action)
	if err != nil {
		return err
	}
	for i := range state.Tasks {
		if state.Tasks[i].ID == task.ID {
			state.Tasks[i] = task
			return SaveTaskState(path, state)
		}
	}
	state.Tasks = append(state.Tasks, task)
	return SaveTaskState(path, state)
}

func MarkTask(path string, site string, action string, taskID string, status string, checkedAt time.Time) error {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	state, err := LoadTaskState(path, site, action)
	if err != nil {
		return err
	}
	for i := range state.Tasks {
		if state.Tasks[i].ID != taskID {
			continue
		}
		if status == "completed" {
			state.Tasks = append(state.Tasks[:i], state.Tasks[i+1:]...)
			return SaveTaskState(path, state)
		}
		state.Tasks[i].Status = status
		state.Tasks[i].LastCheckedAt = checkedAt.Format(time.RFC3339)
		return SaveTaskState(path, state)
	}
	return fmt.Errorf("task %q not found", taskID)
}

func PendingTasks(path string, site string, action string) ([]TaskRecord, error) {
	state, err := LoadTaskState(path, site, action)
	if err != nil {
		return nil, err
	}
	tasks := make([]TaskRecord, len(state.Tasks))
	copy(tasks, state.Tasks)
	return tasks, nil
}

func SortedJobs(jobs []Job) []Job {
	out := make([]Job, len(jobs))
	copy(out, jobs)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Site == out[j].Site {
			if out[i].Action == out[j].Action {
				return out[i].Name < out[j].Name
			}
			return out[i].Action < out[j].Action
		}
		return out[i].Site < out[j].Site
	})
	return out
}

func SelectBatchJobs(jobs []Job, opts ExecuteBatchOptions) ([]Job, error) {
	jobName := strings.TrimSpace(opts.JobName)
	if jobName == "" {
		return SortedJobs(jobs), nil
	}

	var matches []Job
	for _, job := range jobs {
		if job.Name == jobName {
			matches = append(matches, job)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("job %q not found in manifest", jobName)
	case 1:
		return SortedJobs(matches), nil
	default:
		return nil, fmt.Errorf("job %q is ambiguous: %d matching jobs in manifest", jobName, len(matches))
	}
}

func sanitizeName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "-", "/", "-", ":", "-", ".", "-", "_", "-")
	value = replacer.Replace(value)
	value = strings.Trim(value, "-")
	if value == "" {
		return "default"
	}
	return value
}

func ResolvePath(base string, target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return ""
	}
	if filepath.IsAbs(target) {
		return filepath.Clean(target)
	}
	if strings.TrimSpace(base) == "" {
		return filepath.Clean(target)
	}
	return filepath.Clean(filepath.Join(base, target))
}

func executableDir(execPath string) (string, error) {
	if strings.TrimSpace(execPath) == "" {
		return "", fmt.Errorf("executable path is empty")
	}
	if !filepath.IsAbs(execPath) {
		absPath, err := filepath.Abs(execPath)
		if err != nil {
			return "", fmt.Errorf("resolve executable path: %w", err)
		}
		execPath = absPath
	}
	return filepath.Dir(filepath.Clean(execPath)), nil
}

func resolveWithDefault(base string, value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return ResolvePath(base, value)
	}
	return filepath.Clean(fallback)
}

func resolveDir(base string, root string, explicit string, envName string, rootRelative string, systemDefault string) string {
	if strings.TrimSpace(explicit) != "" {
		return ResolvePath(base, explicit)
	}
	if value := strings.TrimSpace(os.Getenv(envName)); value != "" {
		return ResolvePath(base, value)
	}
	if root != "" {
		return filepath.Join(root, rootRelative)
	}
	return filepath.Clean(systemDefault)
}

func resolveFirstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func resolveFirstExisting(override string, candidates ...string) string {
	if override != "" {
		return override
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Clean(candidate)
		}
	}
	for _, candidate := range candidates {
		if candidate != "" {
			return filepath.Clean(candidate)
		}
	}
	return ""
}

func writeJSON(path string, payload interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}
