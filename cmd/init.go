package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultInitSiteName = "default-site"

type initPlan struct {
	ConfigDir   string
	StateDir    string
	LogDir      string
	SitesFile   string
	JobsFile    string
	URLDir      string
	RefreshFile string
	PushFile    string
	SiteName    string
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "create a minimal config skeleton",
	Long: `Create a minimal cdnfix configuration skeleton.

The command creates sites.yaml, jobs.yaml, empty URL files, and the state/log
directories for either a system-install layout or a portable --root layout.
Existing files are preserved unless --force is provided.`,
	Example: `  cdnfix --config-dir /etc/cdnfix --state-dir /var/lib/cdnfix --log-dir /var/log/cdnfix init
  cdnfix --root /opt/cdnfix init
  cdnfix --root /opt/cdnfix --site-name prod-a init`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInit(); err != nil {
			log.Error().Msgf("command init error %s", err.Error())
		}
	},
}

func init() {
	initCmd.Flags().Bool("force", false, "Overwrite generated config and URL files if they already exist")
	initCmd.Flags().String("site-name", defaultInitSiteName, "Site name to use in the generated skeleton")
	_ = viper.BindPFlag("init_force", initCmd.Flags().Lookup("force"))
	_ = viper.BindPFlag("init_site_name", initCmd.Flags().Lookup("site-name"))
	rootCmd.AddCommand(initCmd)
}

func runInit() error {
	plan, err := buildInitPlan(
		viper.GetString("config_dir"),
		viper.GetString("state_dir"),
		viper.GetString("log_dir"),
		viper.GetString("envfile"),
		viper.GetString("manifest"),
		viper.GetString("init_site_name"),
	)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(plan.ConfigDir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(plan.StateDir, "cache"), 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(plan.StateDir, "runs"), 0755); err != nil {
		return fmt.Errorf("create runs dir: %w", err)
	}
	if err := os.MkdirAll(plan.LogDir, 0755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	if err := os.MkdirAll(plan.URLDir, 0755); err != nil {
		return fmt.Errorf("create url dir: %w", err)
	}

	force := viper.GetBool("init_force")
	if err := writeScaffoldFile(plan.SitesFile, renderSitesTemplate(plan.SiteName), force); err != nil {
		return err
	}
	if err := writeScaffoldFile(plan.JobsFile, renderJobsTemplate(plan.SiteName), force); err != nil {
		return err
	}
	if err := writeScaffoldFile(plan.RefreshFile, "", force); err != nil {
		return err
	}
	if err := writeScaffoldFile(plan.PushFile, "", force); err != nil {
		return err
	}

	log.Info().Msgf("initialized config skeleton:\n  sites: %s\n  jobs: %s\n  urls: %s\n  state: %s\n  logs: %s", plan.SitesFile, plan.JobsFile, plan.URLDir, plan.StateDir, plan.LogDir)
	return nil
}

func buildInitPlan(configDir string, stateDir string, logDir string, sitesFile string, jobsFile string, siteName string) (initPlan, error) {
	siteName = strings.TrimSpace(siteName)
	if siteName == "" {
		return initPlan{}, fmt.Errorf("site name is empty")
	}
	configDir = strings.TrimSpace(configDir)
	stateDir = strings.TrimSpace(stateDir)
	logDir = strings.TrimSpace(logDir)
	sitesFile = strings.TrimSpace(sitesFile)
	jobsFile = strings.TrimSpace(jobsFile)
	if configDir == "" || stateDir == "" || logDir == "" || sitesFile == "" || jobsFile == "" {
		return initPlan{}, fmt.Errorf("config layout is not initialized")
	}
	urlDir := filepath.Join(configDir, "urls", siteName)
	return initPlan{
		ConfigDir:   configDir,
		StateDir:    stateDir,
		LogDir:      logDir,
		SitesFile:   sitesFile,
		JobsFile:    jobsFile,
		URLDir:      urlDir,
		RefreshFile: filepath.Join(urlDir, "refresh.txt"),
		PushFile:    filepath.Join(urlDir, "push.txt"),
		SiteName:    siteName,
	}, nil
}

func writeScaffoldFile(path string, content string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("stat %s: %w", path, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create directory for %s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func renderSitesTemplate(siteName string) string {
	return fmt.Sprintf("sites:\n  %s:\n    secret_id: your-secret-id\n    secret_key: your-secret-key\n    region: ap-guangzhou\n", siteName)
}

func renderJobsTemplate(siteName string) string {
	return fmt.Sprintf("jobs:\n  - name: %s-refresh\n    site: %s\n    action: refresh\n    file: ./urls/%s/refresh.txt\n\n  - name: %s-push\n    site: %s\n    action: push\n    file: ./urls/%s/push.txt\n", siteName, siteName, siteName, siteName, siteName, siteName)
}
