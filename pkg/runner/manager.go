/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sas-quantome/interlace-go/pkg/versioner"
)

// maxWarnings specifies the maximum number of warnings allowed before triggering a termination or alert procedure.
const maxWarnings = 33

// MgrBitFlag represents a boolean 16-bit flag for the Manager.
type MgrBitFlag uint16

// constants used in the Manager's bit flag system; aliases are pseudo-variables e.g., const myFlag = MgrFlgAliasA
const (
	MgrFlgHelp MgrBitFlag = 1 << iota
	MgrFlgLoud
	MgrFlgSilent
	MgrFlgReporting
	MgrFlgNoCommit
	MgrFlgNoDelete
	MgrFlgIgnoreWar
	MgrFlgNoErrExit
	MgrFlgNoDownload
	MgrFlgPipeline
	MgrFlgLogOK
	MgrFlgAliasA
	MgrFlgAliasB
	MgrFlgAliasC
	MgrFlgAliasD
	MgrFlgAliasE
)

// Manager encapsulates configurations, flags, and reporting mechanisms for managing application state and behavior.
type Manager struct {
	StartTime time.Time // StartTime represents the starting point in time for the application.
	config    *Framer   // config is a pointer to a Framer used for managing frame-related configurations and operations.
	flag      flagger   // flagger manages bit flags for various operational modes and behaviors.
	report    reporter  // reporter manages logging and reporting mechanisms for the application.
	altLogFp  string    // altLogFp is the path to copy the final log file.
}

// reporter manages logging and reporting mechanisms for the application.
type reporter struct {
	mu     sync.RWMutex // mu protects the following fields
	logFw  *os.File     // logFw is a pointer to the file handle for the log file.
	errLog *log.Logger  // errLog is a pointer to the logger for writing errors to the log file or/and stderr.
	warLog *log.Logger  // warLog is a pointer to the logger for writing warnings to the log file or/and stdout.
	infLog *log.Logger  // infLog is a pointer to the logger for writing informational messages to the log file or/and stdout.
	maxWar int          // maxWar defines the maximum number of warnings that can be logged or tracked by the reporter.
	nWars  int          // nWars is the number of warnings tracked by the reporter.
}

// flagger manages bit flags for various operational modes and behaviors.
type flagger struct {
	mu      sync.RWMutex
	bitFlag BitMask[MgrBitFlag]
}

// NewManager creates a new Manager instance, initializing its fields and starting time.
func NewManager() *Manager {
	mgr := new(Manager)
	mgr.StartTime = time.Now()
	mgr.config = NewFramer()
	mgr.report.maxWar = maxWarnings
	mgr.report.logFw = nil
	return mgr
}

// ConfigureBasicWork gathers basic workflow parameters from the config file and the command line arguments.
func (mgr *Manager) ConfigureBasicWork(conceptD, workflowD, assetU, uriFileU string) {
	// Begin by exposing the boolean flags offered to the user with default values ///////////////////////////////// 1
	help := flag.Bool(ArgHelp())
	silent := flag.Bool(ArgSilent())
	loud := flag.Bool(ArgLoud())
	noDelete := flag.Bool(ArgDelete())
	noDownload := flag.Bool(ArgDownload())
	noCommit := flag.Bool(ArgCommit())
	// Provide the log file parameter option
	logFile := flag.String(ArgLog())
	var basePath, dataPath, configFile, timeTag, baseConcept, workflowID, assetName, uriInpFile string
	// Base path can be already captured from the environment variable ENV_AEE_BASE_PATH.
	if p := mgr.GetConfigString(BasePath); p != "" {
		flag.StringVar(&basePath, "base-path", p, "Agent, event, and entity base path")
	} else {
		flag.StringVar(&basePath, "base-path", "", "Agent, event, and entity base path (required)")
	}
	// Data path can be already captured from the environment variable ENV_ALT_DATA_PATH.
	if p := mgr.GetConfigString(DataPath); p != "" {
		flag.StringVar(&dataPath, "data-path", p, "Data path where to read and download files")
	} else {
		flag.StringVar(&dataPath, "data-path", "", "Data path where to read and download files (required)")
	}
	if mgr.GetBitFlag(MgrFlgPipeline) == false {
		// Event time tag.
		flag.StringVar(&timeTag, "event-time-tag", TimeTagNow(), "Event time tag (required)")
		// Base concept.
		flag.StringVar(&baseConcept, "base-concept", conceptD, "Base concept ontology term to name the entity folder")
		// Workflow ID.
		if workflowD != "" {
			flag.StringVar(&workflowID, "workflow-id", workflowD, "Workflow ID")
		}
		// Asset Name
		if assetU != "" {
			flag.StringVar(&assetName, "asset-name", "", assetU)
		}
		// Config file.
		flag.StringVar(&configFile, "config", "", "Config file (optional)")
	}
	// URI input file
	if uriFileU != "" {
		flag.StringVar(&uriInpFile, "uri-inp-file", "", uriFileU)
	}
	// Parse the command line arguments ///////////////////////////////////////////////////////////////////////////// 2
	flag.Parse()
	// Compute the boolean flags BEFORE starting the loggers.
	mgr.SetBitFlag(MgrFlgHelp, *help)
	mgr.SetBitFlag(MgrFlgSilent, *silent)
	mgr.SetBitFlag(MgrFlgLoud, *loud)
	mgr.SetBitFlag(MgrFlgNoDelete, *noDelete)
	mgr.SetBitFlag(MgrFlgNoDownload, *noDownload)
	mgr.SetBitFlag(MgrFlgNoCommit, *noCommit)
	// Add the log file to the config BEFORE starting the loggers
	mgr.SetConfig(LogFile, *logFile, *logFile != "")
	// Start the stderr and stdout loggers for reporting accordingly to the user configuration above
	mgr.StartLoggers(flag.NFlag())
	// Load the config file if provided: it will override the environment variables.
	if configFile != "" {
		mgr.SetConfig(ConfigFile, configFile, true)
		if err := mgr.LoadConfig(configFile); err != nil {
			mgr.End(err) // exits because err != nil
		}
	}
	// The command line arguments override the existing values except for the time tag.
	mgr.SetConfig(EventTimeTag, timeTag, mgr.GetConfigString(EventTimeTag) == "")
	mgr.SetConfig(DataPath, dataPath, dataPath != "")
	mgr.SetConfig(BasePath, basePath, basePath != "")
	mgr.SetConfig(BaseConcept, baseConcept, baseConcept != "")
	mgr.SetConfig(WorkflowID, workflowID, workflowID != "")
	mgr.SetConfig(AssetNames, assetName, assetName != "")
	mgr.SetConfig(UriInpFile, uriInpFile, uriInpFile != "")
}

// SetResources return eventNS, entityNS, dispenser, URL slice, or err.
func (mgr *Manager) SetResources(assetName string, srcNames []string) (DispenserNamespace, DispenserNamespace, *Dispenser, []string, error) {
	var err error
	var eventNS, entityNS DispenserNamespace
	var url []string
	var base *Dispenser
	eventNS, err = mgr.EventNS()
	if entityNS, err = mgr.EntityNS(assetName); err != nil {
		return eventNS, entityNS, nil, nil, err
	}
	if base, err = mgr.CreateWorkspace([]string{assetName}); err != nil {
		return eventNS, entityNS, nil, nil, err
	}
	if len(srcNames) == 0 || mgr.GetConfigString(UriInpFile) == "" {
		return eventNS, entityNS, base, nil, nil
	}
	if url, err = mgr.SelectFromUriInpAndDownload(srcNames); err != nil {
		return eventNS, entityNS, nil, nil, err
	}
	return eventNS, entityNS, base, url, nil
}

// SelectFromUriInpAndDownload extracts URLs from the UriInpFile and downloads the corresponding URLs if MgrFlgCommit=true.
func (mgr *Manager) SelectFromUriInpAndDownload(selectedFileNames []string) ([]string, error) {
	url, err := mgr.ParseUriInpFile(selectedFileNames)
	if err != nil {
		return nil, err
	}
	if len(url) == 0 {
		return nil, nil
	}
	if mgr.GetBitFlag(MgrFlgNoCommit) == true {
		return nil, nil
	}
	return url, mgr.Download(mgr.GetConfigString(DataPath), url)
}

// ParseUriInpFile returns a slice of URLs when the URL filename matches a srcName, or nil, and
// adds to the manager config full-file-paths that have a fragment key name matching a srcName.
// URI examples:
// https://john.doe@www.example.com:1234/forum/questions/?tag=networking&order=newest#top
// ftp://ftp.is.co.za/rfc/rfc1808.txt
// mailto:John.Doe@example.com
// tel:+1-816-555-1212
// telnet://192.0.2.16:80/
// file:///data/examples/file.txt#ex-inp-file
// File paths must be absolute.
// The file is configured as follows: map[ex-inp-file] = /data/examples/file.txt
// If the ex-inp-file is a fragment key name that matches a name in the srcNames.
// Ignores empty lines or when the first character is #
func (mgr *Manager) ParseUriInpFile(srcNames []string) ([]string, error) {
	lines, err := FileLinesToSlice(mgr.GetConfigString(UriInpFile), true, "#")
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, nil
	}
	url := make([]string, 0)
	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if strings.HasPrefix(line, "http") || strings.HasPrefix(line, "ftp") {
			fn := strings.TrimSpace(path.Base(line))
			if FindItem(srcNames, fn) >= 0 {
				line = strings.TrimSpace(line)
				url = append(url, line)
			}
		} else if after, ok := strings.CutPrefix(line, "file://"); ok {
			s := strings.Split(after, "#")
			if len(s) == 2 {
				key := strings.TrimSpace(s[1])
				if FindItem(srcNames, key) >= 0 {
					mgr.SetConfig(key, s[0], true)
					mgr.Say(fmt.Sprintf("configured file[%s]=%s", key, s[0]))
				}
			}
		}
	}
	return url, nil
}

// LoadConfig loads configuration from the specified file into the manager's configuration.
func (mgr *Manager) LoadConfig(configFile string) error {
	return mgr.config.LoadFrame(configFile)
}

// SetConfig inserts a configuration key-value pair in the manager's configuration if the condition is met.
func (mgr *Manager) SetConfig(key string, value any, conditionOK bool) {
	if conditionOK {
		mgr.config.Insert(key, value)
	}
}

// DelConfig removes a configuration key-value pair from the manager's configuration.
func (mgr *Manager) DelConfig(key string) {
	mgr.config.Delete(key)
}

// GetConfig retrieves a configuration value associated with the provided key from the Manager's config slot.
func (mgr *Manager) GetConfig(key string) any {
	return mgr.config.SearchSlot(key)
}

// GetConfigString retrieves a configuration value associated with the provided key from the Manager's config slot as a string.
func (mgr *Manager) GetConfigString(key string) string {
	return mgr.config.SearchSlotS(key)
}

// GetConfigInt retrieves a configuration value associated with the provided key from the Manager's config slot as an integer.
func (mgr *Manager) GetConfigInt(key string) int {
	return mgr.config.SearchSlotI(key)
}

// PrintConfig writes the configuration to the provided output writer using the specified delimiter.
func (mgr *Manager) PrintConfig(out io.Writer, delimiter string) error {
	return mgr.config.TabWrite(out, delimiter)
}

// SetConfigTopic sets the configuration topic for the application, including module, name, revision, and aim.
func (mgr *Manager) SetConfigTopic(appProject, appName, appAim, appRevision string) {
	mgr.setConfigUser()
	mgr.setConfigEnvVars()
	mgr.config.Insert(AppProject, appProject)
	mgr.config.Insert(AppName, appName)
	mgr.config.Insert(AppRevision, appRevision)
	mgr.config.Insert(AppAim, appAim)
	mgr.config.Insert(AppVersion, versioner.ApiVersionString(appRevision))
}

// CheckConfig verifies the presence of required configuration keys and returns an error if any are missing.
func (mgr *Manager) CheckConfig(key ...string) error {
	var err []error
	if mgr.GetConfigString(AppProject) == "" {
		err = append(err, errors.New("missing application module"))
	}
	if mgr.GetConfigString(AppName) == "" {
		err = append(err, errors.New("missing application name"))
	}
	if mgr.GetConfigString(AppAim) == "" {
		err = append(err, errors.New("missing application aim"))
	}
	if mgr.GetConfigString(AppRevision) == "" {
		err = append(err, errors.New("missing application revision"))
	}
	if p := mgr.GetConfigString(BasePath); p != "" {
		if ValidatePathExists(p) == false {
			err = append(err, errors.New("invalid base path: "+p))
		}
	}
	if p := mgr.GetConfigString(DataPath); p != "" {
		if ValidatePathExists(p) == false {
			err = append(err, errors.New("invalid data path: "+p))
		}
	}
	if p := mgr.GetConfigString(AppsPath); p != "" {
		if ValidatePathExists(p) == false {
			err = append(err, errors.New("invalid apps path: "+p))
		}
	}
	if t := mgr.GetConfigString(EventTimeTag); t != "" {
		if ValidTimeTag(t) == false {
			err = append(err, errors.New("invalid time tag: "+t))
		}
	}
	for _, k := range key {
		if mgr.GetConfig(k) == nil {
			err = append(err, errors.New("missing args for: "+k))
		}
	}
	m := mgr.config.SlotContains("inp-file")
	for i := range m {
		f := mgr.GetConfigString(m[i])
		if ValidatePathExists(f) == false {
			err = append(err, errors.New("invalid input file path: "+f))
		}
	}
	if len(err) > 0 {
		return errors.Join(err...)
	}
	return nil
}

// TaskName generates a unique task name by combining a concept and a term (asset or workflowID) with a colon separator.
func (mgr *Manager) TaskName(term string) string {
	return fmt.Sprintf("%s:%s", mgr.GetConfigString(BaseConcept), term)
}

// SetConfigEnvVars sets configuration values from environment variables for the provided keys.
func (mgr *Manager) SetConfigEnvVars(keys ...string) {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			mgr.SetConfig(key, val, val != "")
		}
	}
}

// SetBitFlag sets or clears a specific bit flag for the Manager based on the provided boolean value.
func (mgr *Manager) SetBitFlag(flag MgrBitFlag, value bool) {
	mgr.flag.mu.Lock()
	defer mgr.flag.mu.Unlock()
	if value {
		mgr.flag.bitFlag.Add(flag)
	} else {
		mgr.flag.bitFlag.Clear(flag)
	}
}

// GetBitFlag returns the current state of a specific bit flag for the Manager.
func (mgr *Manager) GetBitFlag(flag MgrBitFlag) bool {
	mgr.flag.mu.RLock()
	defer mgr.flag.mu.RUnlock()
	return mgr.flag.bitFlag.Has(flag)
}

// StartTimeTag returns the application's StartTime in the format "YYYYMMDDHHMMSS" as the ISO 8601.
func (mgr *Manager) StartTimeTag() string {
	return mgr.StartTime.Format("20060102150405")
}

// ElapsedTime returns the elapsed time between now and the application start formatted as 1h2m0.3s and 300ms.
func (mgr *Manager) ElapsedTime() string {
	return time.Now().Sub(mgr.StartTime).String()
}

// ResetMaxWar sets the maximum war limit to the specified value, ensuring thread-safe modification of the maxWar field.
func (mgr *Manager) ResetMaxWar(n int) {
	mgr.report.mu.Lock()
	defer mgr.report.mu.Unlock()
	mgr.report.maxWar = n
}

// MaxWar flag helper.
func (mgr *Manager) MaxWar() int {
	return mgr.report.maxWar
}

// ArgHelp flag helper.
func ArgHelp() (string, bool, string) {
	return "help", false, "Displays this help page"
}

// ArgSilent flag helper.
func ArgSilent() (string, bool, string) {
	return "silent", false, "Suppresses all standard output messages"
}

// ArgLoud flag helper.
func ArgLoud() (string, bool, string) {
	return "loud", false, "Writes additional information if available"
}

// ArgDelete flag helper.
func ArgDelete() (string, bool, string) {
	return "no-delete", false, "Disables deleting the artifact files"
}

// ArgDownload flag helper.
func ArgDownload() (string, bool, string) {
	return "no-download", false, "Disables downloading data when necessary"
}

// ArgCommit flag helper.
func ArgCommit() (string, bool, string) {
	return "no-commit", false, "Disables committing the work"
}

// ArgLog flag helper.
func ArgLog() (string, string, string) {
	return "log", "", "Logfile path to write stderr & stdout regardless of the -silent flag (default no log file)"
}

// PrintStartingInfo prints the application's starting information, including the application name, version, module, aim, and start time tag.
func (mgr *Manager) PrintStartingInfo() {
	mgr.Say(fmt.Sprintf("application name: %s", mgr.GetConfigString(AppName)))
	mgr.Say(fmt.Sprintf("application version: %s", mgr.GetConfigString(AppVersion)))
	mgr.Say(fmt.Sprintf("application module: %s", mgr.GetConfigString(AppProject)))
	mgr.Say(fmt.Sprintf("application aim: %s", strings.Split(mgr.GetConfigString(AppAim), "\n")[0]))
	mgr.Say(fmt.Sprintf("application start time tag: %s", mgr.StartTimeTag()))
	mgr.Say("listing configuration variables ...")
	mgr.config.Log(mgr)
}

// HelpPage generates and prints a formatted help page with application details and optional documentation.
func (mgr *Manager) HelpPage(doc string) {
	var buf bytes.Buffer
	buf.WriteString("\nProject: ")
	buf.WriteString(mgr.GetConfigString(AppProject))
	buf.WriteString("\nAppName: ")
	buf.WriteString(mgr.GetConfigString(AppName))
	buf.WriteString("\nVersion: ")
	buf.WriteString(mgr.GetConfigString(AppVersion))
	buf.WriteString("\nPurpose: ")
	buf.WriteString(mgr.GetConfigString(AppAim))
	if doc != "" {
		buf.WriteString(fmt.Sprintf("\n\n%sOptions:\n\n", doc))
	} else {
		buf.WriteString("\n\nOptions:\n\n")
	}
	if _, err := os.Stderr.WriteString(buf.String()); err != nil {
		panic(err)
	}
	flag.PrintDefaults()
	if _, err := os.Stderr.WriteString("\n"); err != nil {
		panic(err)
	}
}

// Say prints a message to the standard output and/or log file.
func (mgr *Manager) Say(inf string) {
	if (!mgr.GetBitFlag(MgrFlgSilent) || mgr.GetBitFlag(MgrFlgLogOK)) && mgr.AssureReporting(inf) {
		mgr.logInf(inf)
	}
}

// War prints a warning message to the standard output and/or log file.
func (mgr *Manager) War(msg string) {
	if mgr.AssureReporting(msg) {
		mgr.logWar(msg)
	}
	if mgr.WarningsExceedLimit() {
		w := fmt.Sprintf("too many warnings: %d/%d: exiting program", mgr.report.nWars, mgr.report.maxWar)
		mgr.End(errors.New(w))
	}
}

// End closes the log file and exits the program with an error message if the error is not nil.
func (mgr *Manager) End(err error) {
	if mgr.GetBitFlag(MgrFlgHelp) == true {
		return
	}
	mgr.Say("number of warnings caught: " + strconv.Itoa(mgr.TotalWarnings()))
	if err != nil {
		if mgr.AssureReporting(err.Error()) {
			mgr.logErr(err)
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "err: %s\n", err.Error())
		}
		mgr.Say(fmt.Sprintf("elapsed time just before closing: %s", mgr.ElapsedTime()))
		mgr.Say("end of process: incomplete")
		mgr.closeResources()
		if mgr.GetBitFlag(MgrFlgNoErrExit) == true {
			return // you must have a very good reason to do this
		}
		os.Exit(1)
	}
	mgr.Say(fmt.Sprintf("elapsed time just before closing: %s", mgr.ElapsedTime()))
	mgr.Say("end of process: complete")
	mgr.closeResources()
}

// WarningsExceedLimit returns true if the number of warnings exceeds the maximum allowed.
func (mgr *Manager) WarningsExceedLimit() bool {
	mgr.report.mu.RLock()
	defer mgr.report.mu.RUnlock()
	return !mgr.GetBitFlag(MgrFlgIgnoreWar) && !mgr.GetBitFlag(MgrFlgPipeline) && mgr.report.nWars > mgr.report.maxWar
}

// TotalWarnings returns the total number of warnings recorded in the report.
func (mgr *Manager) TotalWarnings() int {
	mgr.report.mu.RLock()
	defer mgr.report.mu.RUnlock()
	return mgr.report.nWars
}

// AssureReporting checks if reporting is enabled and the message is not empty.
func (mgr *Manager) AssureReporting(msg string) bool {
	return mgr.GetBitFlag(MgrFlgReporting) && msg != ""
}

// StartLoggers starts the logging system: log file, stdout, and stderr, depending on the user arguments.
// If the flags.nFlag == 0 assumes the help page it is requested, and it sets the Help bit flag true.
func (mgr *Manager) StartLoggers(nFlag int) {
	if mgr.GetBitFlag(MgrFlgReporting) == false {
		mgr.startLogging(mgr.GetBitFlag(MgrFlgSilent), mgr.GetConfigString(LogFile))
	}
	if mgr.GetBitFlag(MgrFlgHelp) == true {
		return
	}
	if nFlag == 0 {
		mgr.SetBitFlag(MgrFlgHelp, true)
		return
	}
}

// EventNS generates a DispenserNamespace using configuration values for concept, workflow ID, and event time tag.
func (mgr *Manager) EventNS() (DispenserNamespace, error) {
	concept := mgr.GetConfigString(BaseConcept)
	workflowID := mgr.GetConfigString(WorkflowID)
	timeTag := mgr.GetConfigString(EventTimeTag)
	return DispenseNamespace(concept, workflowID, timeTag, true)
}

// EntityNS generates a DispenserNamespace using configuration values for concept, asset, and event time tag.
func (mgr *Manager) EntityNS(asset string) (DispenserNamespace, error) {
	concept := mgr.GetConfigString(BaseConcept)
	timeTag := mgr.GetConfigString(EventTimeTag)
	return DispenseNamespace(concept, asset, timeTag, false)
}

// CreateWorkspace initializes a base workspace with event and entity directories based on provided assets.
func (mgr *Manager) CreateWorkspace(assets []string) (*Dispenser, error) {
	var err error
	var ns DispenserNamespace
	concept := mgr.GetConfigString(BaseConcept)
	workflowID := mgr.GetConfigString(WorkflowID)
	timeTag := mgr.GetConfigString(EventTimeTag)
	// Create the base dispenser.
	base := NewDispenser(mgr.GetConfigString(BasePath))
	// Create the event workspace.
	err = base.AddEvent(concept, workflowID, timeTag)
	if err != nil {
		return nil, err
	}
	if ns, err = DispenseNamespace(concept, workflowID, timeTag, true); err != nil {
		return nil, err
	}
	if mgr.GetBitFlag(MgrFlgNoCommit) == false {
		if err = base.CreateEventDirectories(ns); err != nil {
			return nil, err
		}
	}
	// Create the entity workspaces.
	for _, asset := range assets {
		if err = base.AddEntity(concept, asset, timeTag); err != nil {
			return nil, err
		}
		if ns, err = DispenseNamespace(concept, asset, timeTag, false); err != nil {
			return nil, err
		}
		if mgr.GetBitFlag(MgrFlgNoCommit) == false {
			if err = base.CreateEntityDirectories(ns); err != nil {
				return nil, err
			}
		}
	}
	return base, nil
}

func (mgr *Manager) startLogging(quiet bool, logFile string) {
	mgr.report.mu.Lock()
	defer mgr.report.mu.Unlock()
	if logFile != "" {
		var err error
		mgr.report.logFw, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			mgr.End(err) // will exit because err != nil
		}
		mgr.SetBitFlag(MgrFlgLogOK, true)
		if quiet == true {
			// Log file and stderr only.
			mgr.report.infLog = log.New(mgr.report.logFw, "Inf: ", log.Ldate|log.Lmicroseconds)
			mgr.report.warLog = log.New(mgr.report.logFw, "War: ", log.Ldate|log.Lmicroseconds)
			mgr.report.errLog = log.New(io.MultiWriter(mgr.report.logFw, os.Stderr), "Err: ", log.Ldate|log.Lmicroseconds)
		} else {
			// Log file and stdout and stderr.
			mgr.report.infLog = log.New(io.MultiWriter(mgr.report.logFw, os.Stdout), "Inf: ", log.Ldate|log.Lmicroseconds)
			mgr.report.warLog = log.New(io.MultiWriter(mgr.report.logFw, os.Stdout), "War: ", log.Ldate|log.Lmicroseconds)
			mgr.report.errLog = log.New(io.MultiWriter(mgr.report.logFw, os.Stderr), "Err: ", log.Ldate|log.Lmicroseconds)
		}
	} else {
		if quiet == true {
			// Stderr only.
			mgr.report.warLog = log.New(os.Stderr, "War: ", log.Ldate|log.Lmicroseconds)
			mgr.report.errLog = log.New(os.Stderr, "Err: ", log.Ldate|log.Lmicroseconds)
		} else {
			// Stdout and stderr only.
			mgr.report.infLog = log.New(os.Stdout, "Inf: ", log.Ldate|log.Lmicroseconds)
			mgr.report.warLog = log.New(os.Stdout, "War: ", log.Ldate|log.Lmicroseconds)
			mgr.report.errLog = log.New(os.Stderr, "Err: ", log.Ldate|log.Lmicroseconds)
		}
	}
	mgr.SetBitFlag(MgrFlgReporting, true)
}

func (mgr *Manager) logInf(inf string) {
	mgr.report.mu.Lock()
	defer mgr.report.mu.Unlock()
	mgr.report.infLog.Println(inf)
}

func (mgr *Manager) logWar(inf string) {
	mgr.report.mu.Lock()
	defer mgr.report.mu.Unlock()
	mgr.report.warLog.Println(inf)
	mgr.report.nWars++
}

func (mgr *Manager) logErr(err error) {
	mgr.report.mu.Lock()
	defer mgr.report.mu.Unlock()
	mgr.report.errLog.Println(err.Error())
}

func (mgr *Manager) closeResources() {
	mgr.report.mu.RLock()
	defer mgr.report.mu.RUnlock()
	if mgr.report.logFw != nil {
		if err := mgr.report.logFw.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "err: %s", err.Error())
		} else if mgr.altLogFp != "" {
			if e := CopyFile(mgr.report.logFw.Name(), mgr.altLogFp); e != nil {
				_, _ = fmt.Fprintf(os.Stderr, "err: copying log file: %s", e.Error())
			}
		}
	}
}

func (mgr *Manager) setConfigUser() {
	u, err := user.Current()
	if err != nil {
		return
	}
	mgr.SetConfig(UserLogin, u.Username, u.Username != "")
	mgr.SetConfig(UserName, u.Name, u.Name != "")
	mgr.SetConfig(UserHomeDir, u.HomeDir, u.HomeDir != "")
}

func (mgr *Manager) setConfigEnvVars() {
	var v []string
	v = append(v, EnvBasePath, EnvDataPath, EnvAppsPath)
	for i := range v {
		if mgr.config.SearchSlotS(v[i]) == "" {
			p := os.Getenv(v[i])
			if p != "" {
				mgr.SetConfig(v[i], p, p != "")
				switch v[i] {
				case EnvBasePath:
					mgr.SetConfig(BasePath, p, p != "")
				case EnvDataPath:
					mgr.SetConfig(DataPath, p, p != "")
				case EnvAppsPath:
					mgr.SetConfig(AppsPath, p, p != "")
				}
			}
		}
	}
}

// ParseJobLogFile reads a job log file, evaluates its contents, and reports warnings, errors, or failures as needed.
func (mgr *Manager) ParseJobLogFile(fp string) error {
	if mgr.GetBitFlag(MgrFlgNoCommit) == true {
		return nil
	}
	file, err := os.Open(fp)
	if err != nil {
		mgr.Say(fmt.Sprintf("failed opening: %s: %s", fp, err.Error()))
		return fmt.Errorf("failure reading: %s", fp)
	}
	scanner := bufio.NewScanner(file)
	failed := false
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "War: 2") {
			mgr.Say(fmt.Sprintf("war in: %s: %s", fp, scanner.Text()))
		} else if strings.HasPrefix(scanner.Text(), "Err: 2") {
			mgr.Say(fmt.Sprintf("err in: %s: %s", fp, scanner.Text()))
			err = fmt.Errorf("job failed: %s", path.Base(fp))
			break
		} else if ErrorInLine(scanner.Text()) {
			mgr.Say(fmt.Sprintf("err in: %s: %s", fp, scanner.Text()))
			err = fmt.Errorf("job failed: %s", path.Base(fp))
			break
		} else if strings.Contains(scanner.Text(), "application date time tag: ") {
			failed = true
		} else if failed && strings.Contains(scanner.Text(), "end of process: complete") {
			failed = false
			break
		}
	}
	if failed {
		mgr.War(fmt.Sprintf("end of process not completed: %s", fp))
	}
	if err != nil {
		failed = true
	}
	if err = file.Close(); err != nil {
		failed = true
		mgr.War(fmt.Sprintf("failed to close file: %s", fp))
	}
	if scanner.Err() != nil {
		failed = true
		mgr.War(fmt.Sprintf("failed scanning: %s: %v", fp, scanner.Err()))
	}
	if failed {
		return fmt.Errorf("see: %s", fp)
	}
	return nil
}

// DataPathInpFile returns DATA_PATH/filename.
func (mgr *Manager) DataPathInpFile(name string) string {
	return path.Join(mgr.GetConfigString(DataPath), name)
}

// ConfigureWork gathers the application parameters from the config file and the command line arguments.
// If any input path is invalid, or a required parameter is missing, it exits explaining all errors.
func (mgr *Manager) ConfigureWork(concept, workflowID string, hasURI, askForAsset bool) {
	var err error
	// Validate the expected configuration: BasePath, DataPath, AppsPath, and TimeTag are always checked.
	// Pass the WorkflowID to the CheckConfig expected from ConfigureBasicWork.
	if hasURI {
		urlUsage := "URI list of URL, mailto, telnet, absolute-file-path#key (unique key as URI fragment), etc."
		if askForAsset {
			mgr.ConfigureBasicWork(concept, workflowID, "lowercased asset name comma-delimited if more than one", urlUsage)
			err = mgr.CheckConfig(WorkflowID, UriInpFile, AssetNames, BaseConcept)
		} else {
			mgr.ConfigureBasicWork(concept, workflowID, "", urlUsage)
			if mgr.GetBitFlag(MgrFlgPipeline) == true {
				err = mgr.CheckConfig(UriInpFile)
			} else {
				err = mgr.CheckConfig(WorkflowID, UriInpFile)
			}
		}
	} else {
		mgr.ConfigureBasicWork(concept, workflowID, "", "")
		err = mgr.CheckConfig(WorkflowID)
	}
	if err != nil {
		mgr.End(err)
	}
}

// CopyLogFile sets the alternate log file path for copying the final log file.
func (mgr *Manager) CopyLogFile(fp string) {
	mgr.altLogFp = fp
}
