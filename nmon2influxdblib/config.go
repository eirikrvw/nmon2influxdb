// nmon2influxdb
// author: adejoux@djouxtech.net

package nmon2influxdblib

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/adejoux/influxdbclient"
	"github.com/codegangsta/cli"
	"github.com/naoina/toml"
)

// Config is the configuration structure used by nmon2influxdb
type Config struct {
	Debug                bool
	Timezone             string
	InfluxdbUser         string
	InfluxdbPassword     string
	InfluxdbServer       string
	InfluxdbPort         string
	InfluxdbDatabase     string
	GrafanaUser          string
	GrafanaPassword      string
	GrafanaURL           string `toml:"grafana_URL"`
	GrafanaAccess        string
	GrafanaDatasource    string
	HMCServer            string `toml:"hmc_server"`
	HMCUser              string `toml:"hmc_user"`
	HMCPassword          string `toml:"hmc_password"`
	HMCDatabase          string `toml:"hmc_database"`
	HMCDataRetention     string `toml:"hmc_data_retention"`
	HMCManagedSystem     string `toml:"hmc_managed_system"`
	HMCManagedSystemOnly bool   `toml:"hmc_managed_system_only"`
	HMCSamples           int    `toml:"hmc_samples"`
	ImportSkipDisks      bool
	ImportAllCpus        bool
	ImportBuildDashboard bool
	ImportForce          bool
	ImportSkipMetrics    string
	ImportLogDatabase    string
	ImportLogRetention   string
	ImportDataRetention  string
	ImportSSHUser        string `toml:"import_ssh_user"`
	ImportSSHKey         string `toml:"import_ssh_key"`
	DashboardWriteFile   bool
	StatsLimit           int
	StatsSort            string
	StatsFilter          string
	StatsFrom            string
	StatsTo              string
	StatsHost            string
	Metric               string `toml:"metric,omitempty"`
	ListFilter           string `toml:",omitempty"`
	ListHost             string `toml:",omitempty"`
	Inputs               Inputs `toml:"input"`
}

// Inputs allows to put multiple input in the configuration file
type Inputs []Input

// Input specify how to apply new filters
type Input struct {
	Measurement string
	Name        string
	Match       string
	Tags        Tags `toml:"tag"`
}

// InitConfig setup initial configuration with sane values
func InitConfig() Config {
	currUser, _ := user.Current()
	home := currUser.HomeDir
	sshKey := filepath.Join(home, "/.ssh/id_rsa")

	return Config{Debug: false,
		Timezone:             "Europe/Paris",
		InfluxdbUser:         "root",
		InfluxdbPassword:     "root",
		InfluxdbServer:       "localhost",
		InfluxdbPort:         "8086",
		InfluxdbDatabase:     "nmon_reports",
		HMCUser:              "hscroot",
		HMCPassword:          "abc123",
		HMCDatabase:          "nmon2influxdbHMC",
		GrafanaUser:          "admin",
		GrafanaPassword:      "admin",
		GrafanaURL:           "http://localhost:3000",
		GrafanaAccess:        "direct",
		GrafanaDatasource:    "nmon2influxdb",
		ImportSkipDisks:      false,
		ImportAllCpus:        false,
		ImportBuildDashboard: false,
		ImportForce:          false,
		ImportLogDatabase:    "nmon2influxdb_log",
		ImportLogRetention:   "2d",
		ImportSSHUser:        currUser.Username,
		ImportSSHKey:         sshKey,
		DashboardWriteFile:   false,
		ImportSkipMetrics:    "JFSINODE|TOP|PCPU",
		StatsLimit:           20,
		StatsSort:            "mean",
		StatsFilter:          "",
		StatsFrom:            "",
		StatsTo:              "",
		StatsHost:            "",
	}
}

//GetCfgFile returns the current configuration file path
func GetCfgFile() string {
	// if configuration file exist in /etc/nmon2influxdb. Stop here.
	if IsFile("/etc/nmon2influxdb/nmon2influxdb.cfg") {
		return "/etc/nmon2influxdb/nmon2influxdb.cfg"
	}

	currUser, _ := user.Current()
	home := currUser.HomeDir
	homeCFGfile := filepath.Join(home, ".nmon2influxdb.cfg")
	return homeCFGfile
}

//IsFile returns true if the file doesn't exist
func IsFile(file string) bool {
	stat, err := os.Stat(file)
	if err != nil {
		return false
	}
	if stat.Mode().IsRegular() {
		return true
	}

	return false
}

//BuildCfgFile creates a default configuration file
func (config *Config) BuildCfgFile(cfgfile string) {
	file, err := os.Create(cfgfile)
	CheckError(err)
	defer file.Close()
	writer := bufio.NewWriter(file)
	b, err := toml.Marshal(*config)
	CheckError(err)
	r := bytes.NewReader(b)
	r.WriteTo(writer)
	writer.Flush()
	fmt.Printf("Generating default configuration file : %s\n", cfgfile)
}

// LoadCfgFile loads current configuration file settings
func (config *Config) LoadCfgFile() (cfgfile string) {

	cfgfile = GetCfgFile()

	//it would be only if no conf file exists. And it will build a configuration file in the home directory
	if !IsFile(cfgfile) {
		config.BuildCfgFile(cfgfile)
	}

	file, err := os.Open(cfgfile)
	if err != nil {
		fmt.Printf("Error opening configuration file %s\n", cfgfile)
		return
	}

	defer file.Close()
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		CheckError(err)
	}

	if err := toml.Unmarshal(buf, &config); err != nil {
		fmt.Printf("syntax error in configuration file: %s \n", err.Error())
		os.Exit(1)
	}
	return
}

// AddDashboardParams initialize default parameters for dashboard
func (config *Config) AddDashboardParams() {
	dfltConfig := InitConfig()
	dfltConfig.LoadCfgFile()

	config.GrafanaAccess = dfltConfig.GrafanaAccess
	config.GrafanaURL = dfltConfig.GrafanaURL
	config.GrafanaDatasource = dfltConfig.GrafanaDatasource
	config.GrafanaUser = dfltConfig.GrafanaUser
	config.GrafanaPassword = dfltConfig.GrafanaPassword
}

// ParseParameters parse parameter from command line in Config struct
func ParseParameters(c *cli.Context) (config *Config) {
	config = new(Config)
	*config = InitConfig()
	config.LoadCfgFile()

	config.Metric = c.String("metric")
	config.StatsHost = c.String("statshost")
	config.StatsFrom = c.String("from")
	config.StatsTo = c.String("to")
	config.StatsLimit = c.Int("limit")
	config.StatsFilter = c.String("filter")
	config.ImportSkipDisks = c.Bool("nodisks")
	if c.IsSet("cpus") {
		config.ImportAllCpus = c.Bool("cpus")
	}
	config.ImportBuildDashboard = c.Bool("build")
	config.ImportSkipMetrics = c.String("skip_metrics")
	config.ImportLogDatabase = c.String("log_database")
	config.ImportLogRetention = c.String("log_retention")
	config.DashboardWriteFile = c.Bool("file")
	config.ListFilter = c.String("filter")
	config.ImportForce = c.Bool("force")
	config.ListHost = c.String("host")
	config.GrafanaUser = c.String("guser")
	config.GrafanaPassword = c.String("gpassword")
	config.GrafanaAccess = c.String("gaccess")
	config.GrafanaURL = c.String("gurl")
	config.GrafanaDatasource = c.String("datasource")
	config.Debug = c.GlobalBool("debug")
	config.HMCServer = c.String("hmc")
	config.HMCUser = c.String("hmcuser")
	config.HMCPassword = c.String("hmcpass")
	config.HMCManagedSystem = c.String("managed_system")
	config.HMCManagedSystemOnly = c.Bool("managed_system-only")
	config.HMCSamples = c.Int("samples")
	config.InfluxdbServer = c.GlobalString("server")
	config.InfluxdbUser = c.GlobalString("user")
	config.InfluxdbPort = c.GlobalString("port")
	config.InfluxdbDatabase = c.GlobalString("db")
	config.InfluxdbPassword = c.GlobalString("pass")
	config.Timezone = c.GlobalString("tz")

	if config.ImportBuildDashboard {
		config.AddDashboardParams()
	}

	return

}

// ConnectDB connect to the specified influxdb database
func (config *Config) ConnectDB(db string) *influxdbclient.InfluxDB {
	influxdbConfig := influxdbclient.InfluxDBConfig{
		Host:     config.InfluxdbServer,
		Port:     config.InfluxdbPort,
		Database: db,
		User:     config.InfluxdbUser,
		Pass:     config.InfluxdbPassword,
		Debug:    config.Debug,
	}

	influxdb, err := influxdbclient.NewInfluxDB(influxdbConfig)
	CheckError(err)

	return &influxdb
}

// GetDB create or get the influxdb database used for nmon data
func (config *Config) GetDB(dbType string) *influxdbclient.InfluxDB {

	db := config.InfluxdbDatabase
	retention := config.ImportDataRetention

	if dbType == "hmc" {
		db = config.HMCDatabase
		retention = config.HMCDataRetention
	}

	influxdb := config.ConnectDB(db)

	if exist, _ := influxdb.ExistDB(db); exist != true {
		fmt.Printf("Creating InfluxDB database %s\n", db)
		_, createErr := influxdb.CreateDB(db)
		CheckError(createErr)
	}

	// update default retention policy if ImportDataRetention is set
	if len(retention) > 0 {
		// Get default retention policy name
		policyName, policyErr := influxdb.GetDefaultRetentionPolicy()
		CheckError(policyErr)
		fmt.Printf("Updating  %s retention policy to keep only the last %s days. Timestamp based.\n", policyName, retention)
		_, err := influxdb.UpdateRetentionPolicy(policyName, retention, true)
		CheckError(err)
	}
	return influxdb
}

// GetLogDB create or get the influxdb database like defined in config
func (config *Config) GetLogDB() *influxdbclient.InfluxDB {

	influxdb := config.ConnectDB(config.ImportLogDatabase)

	if exist, _ := influxdb.ExistDB(config.ImportLogDatabase); exist != true {
		_, err := influxdb.CreateDB(config.ImportLogDatabase)
		CheckError(err)
		_, err = influxdb.SetRetentionPolicy("log_retention", config.ImportLogRetention, true)
		CheckError(err)
	} else {
		_, err := influxdb.UpdateRetentionPolicy("log_retention", config.ImportLogRetention, true)
		CheckError(err)
	}
	return influxdb
}
