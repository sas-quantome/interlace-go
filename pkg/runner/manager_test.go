/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"errors"
	"flag"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/sas-quantome/interlace-go/pkg/tester"
	"github.com/sas-quantome/interlace-go/pkg/versioner"
	"github.com/stretchr/testify/require"
)

func TestManager_NewManager(t *testing.T) {
	mgr := NewManager()
	require.NotNil(t, mgr.config)
	require.NotNil(t, mgr.StartTime)
	require.Equal(t, 0, mgr.TotalWarnings())
	require.Equal(t, maxWarnings, mgr.MaxWar())
}

func TestManager_ResetMaxWar(t *testing.T) {
	mgr := NewManager()
	mgr.ResetMaxWar(666)
	require.Equal(t, 666, mgr.MaxWar())
}

func TestManager_WarningsExceedLimit(t *testing.T) {
	mgr := NewManager()
	require.False(t, mgr.WarningsExceedLimit())
	mgr.ResetMaxWar(1)
	// Warnings are incremented with the logger: War()
	mgr.StartLoggers(-1)
	mgr.War("this is warning 1")
	mgr.SetBitFlag(MgrFlgNoErrExit, true) // otherwise this test will fail
	mgr.War("this is warning 2")          // with the MgrFlgNoErrExit flag, the End() called by War() will not os.Exit(1)
}

func TestManager_SetConfigEnvVars(t *testing.T) {
	mgr := NewManager()
	// CheckConfig will look for the topic parameters
	mgr.SetConfigTopic(
		"age_reversal",
		"find_senescence",
		"Sets the config environment variables",
		"1")
	require.NoError(t, mgr.CheckConfig(UserHomeDir))
	// Set two environment variables normally found in all systems
	mgr.SetConfigEnvVars(UserHomeDir, "PATH")
	// Now it should pass
	require.NoError(t, mgr.CheckConfig(UserHomeDir, "PATH"))
}

func TestManager_SetConfigTopic(t *testing.T) {
	mgr := NewManager()
	// CheckConfig will look for the topic parameters that are expected for every application
	mgr.SetConfigTopic(
		"age_reversal",
		"find_senescence",
		"Checks for the topic parameters",
		"1")
	require.Equal(t, "age_reversal", mgr.GetConfigString(AppProject))
	require.Equal(t, "find_senescence", mgr.GetConfigString(AppName))
	require.Equal(t, "Checks for the topic parameters", mgr.GetConfigString(AppAim))
	require.Equal(t, versioner.ApiVersion+"1", mgr.GetConfigString(AppVersion))
	require.NoError(t, mgr.CheckConfig(AppProject, AppName, AppAim, AppVersion))
}

func TestManager_LoadConfig(t *testing.T) {
	mgr := NewManager()
	mgr.SetConfigTopic(
		"age_reversal",
		"find_senescence",
		"Reads the config file",
		"1")
	require.NoError(t, mgr.LoadConfig(path.Join(tester.TestDataPath(), "config.tsv")))
	// Assume the application requires the following parameters letting the others optional
	require.NoError(t, mgr.CheckConfig("activity", "workflow", "pipeline", AppName))
}

func TestManager_CheckConfig(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{
		oldArgs[0],
		"-cid", "abc",
		"-n", "10",
		"-i", path.Join(tester.TestDataPath(), "senescence.tsv"),
	}
	// Start the manager
	mgr := NewManager()
	// Declare the hidden topic parameters from the user that are required for the application
	mgr.SetConfigTopic(
		"age_reversal",
		"find_senescence",
		"Looks for the TSV senescence file and prints the config parameters",
		"1")
	// Expose the boolean flags offered to the user with default values
	help := flag.Bool(ArgHelp())
	silent := flag.Bool(ArgSilent())
	loud := flag.Bool(ArgLoud())
	noDelete := flag.Bool(ArgDelete())
	noDownload := flag.Bool(ArgDownload())
	noCommit := flag.Bool(ArgCommit())
	// Expose the config parameters that the user can provide
	var cid, iFn string
	var num int
	flag.StringVar(&cid, "cid", "", "Content ID")
	flag.StringVar(&iFn, "i", "", "Input file")
	flag.IntVar(&num, "n", 0, "Number of rows to read")
	// Provide the log file parameter option
	log := flag.String(ArgLog())
	// Parse the command line arguments
	flag.Parse()
	// Collect the configuration parameters (false conditions are ignored)
	mgr.SetConfig("cid", cid, cid != "")
	mgr.SetConfig("num", num, num >= 0)
	mgr.SetConfig("iFn", iFn, iFn != "")
	// Handle the log file parameter with the controlled vocabulary config key name
	mgr.SetConfig(LogFile, *log, *log != "")
	// Compute the boolean flags using the controlled vocabulary bit flag key name
	mgr.SetBitFlag(MgrFlgHelp, *help)
	mgr.SetBitFlag(MgrFlgSilent, *silent)
	mgr.SetBitFlag(MgrFlgLoud, *loud)
	mgr.SetBitFlag(MgrFlgNoDelete, *noDelete)
	mgr.SetBitFlag(MgrFlgNoDownload, *noDownload)
	mgr.SetBitFlag(MgrFlgNoCommit, *noCommit)
	// Check the configuration parameters that are expected beside the application parameters
	require.NoError(t, mgr.CheckConfig("cid", "num", "iFn"))
	// The LogFile is not expected in this test
	require.Error(t, mgr.CheckConfig("cid", "num", "iFn", LogFile))
	// Deleting the application name from the configuration makes the config test fail
	mgr.DelConfig(AppName)
	require.Error(t, mgr.CheckConfig("cid", "num", "iFn"))
	// Start the stderr and stdout loggers for reporting accordingly to the user configuration
	mgr.StartLoggers(flag.NFlag())
	// If the user requested the help page or number of command line arguments is zero, show usage, and exit
	if mgr.GetBitFlag(MgrFlgHelp) == true {
		return // show the help page
	}
	// Commit the work
	if mgr.GetBitFlag(MgrFlgNoCommit) == false {
		// Look for the input file and printout the configuration parameters
		require.True(t, ValidatePathExists(mgr.GetConfigString("iFn")))
		require.NoError(t, mgr.PrintConfig(os.Stdout, "\t"))
		t.Log("Work is committed!")
	}
}

func TestManager_HelpPage(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{oldArgs[0], "-help"}
	mgr := NewManager()
	mgr.SetConfigTopic(
		"age_reversal",
		"find_senescence",
		"Shows the help page",
		"1")
	help := flag.Bool(ArgHelp())
	silent := flag.Bool(ArgSilent())
	noDownload := flag.Bool(ArgDownload())
	noCommit := flag.Bool(ArgCommit())
	loud := flag.Bool(ArgLoud())
	noDelete := flag.Bool(ArgDelete())
	flag.Parse()
	mgr.SetBitFlag(MgrFlgNoDownload, *noDownload)
	mgr.SetBitFlag(MgrFlgNoCommit, *noCommit)
	mgr.SetBitFlag(MgrFlgHelp, *help)
	mgr.SetBitFlag(MgrFlgSilent, *silent)
	mgr.SetBitFlag(MgrFlgLoud, *loud)
	mgr.SetBitFlag(MgrFlgNoDelete, *noDelete)
	mgr.StartLoggers(flag.NFlag())
	if mgr.GetBitFlag(MgrFlgHelp) == true {
		mgr.HelpPage("Bla bla bla bla.\n\n")
		return
	}
	require.Fail(t, "not supposed to get here")
}

func TestManager_StartTimeTag(t *testing.T) {
	mgr := NewManager()
	require.True(t, ValidTimeTag(mgr.StartTimeTag()))
}

func TestManager_ElapsedTime(t *testing.T) {
	mgr := NewManager()
	time.Sleep(time.Second)
	require.True(t, strings.HasPrefix(mgr.ElapsedTime(), "1.0"))
}

func TestManager_Say(t *testing.T) {
	mgr := NewManager()
	mgr.StartLoggers(-1)
	mgr.Say("testing Say() 1")
	mgr.Say("testing Say() 2")
	mgr.Say("testing Say() 3")
	mgr.Say("testing Say() 4")
	require.Equal(t, 0, mgr.TotalWarnings())
}

func TestManager_War(t *testing.T) {
	mgr := NewManager()
	mgr.StartLoggers(-1)
	mgr.War("testing War() 1")
	mgr.War("testing War() 2")
	mgr.War("testing War() 3")
	mgr.War("testing War() 4")
	require.Equal(t, 4, mgr.TotalWarnings())
}

func TestManager_End(t *testing.T) {
	// End without an error
	mgr := NewManager()
	mgr.StartLoggers(-1)
	mgr.End(nil)
	// End with an error
	mgr = NewManager()
	mgr.StartLoggers(-1)
	mgr.SetBitFlag(MgrFlgNoErrExit, true) // otherwise this test will fail
	mgr.End(errors.New("testing End()"))  // with the MgrFlgNoErrExit flag, the End() will not os.Exit(1)
	require.Equal(t, 0, mgr.TotalWarnings())
	// End with a warning exceeding the limit
	mgr = NewManager()
	mgr.StartLoggers(-1)
	mgr.ResetMaxWar(0)                    // zero tolerance
	mgr.SetBitFlag(MgrFlgNoErrExit, true) // otherwise this test will fail
	mgr.War("no warnings allowed")        // with the MgrFlgNoErrExit flag, the End() called by War() will not os.Exit(1)
	require.Equal(t, 1, mgr.TotalWarnings())
}

func TestManager_BitAliases(t *testing.T) {
	a := MgrFlgAliasA
	b := MgrFlgAliasB
	c := MgrFlgAliasC
	d := MgrFlgAliasD
	e := MgrFlgAliasE
	mgr := NewManager()
	mgr.SetBitFlag(a, true)
	mgr.SetBitFlag(b, true)
	mgr.SetBitFlag(c, true)
	mgr.SetBitFlag(d, true)
	mgr.SetBitFlag(e, true)
	require.True(t, mgr.GetBitFlag(a))
	require.True(t, mgr.GetBitFlag(b))
	require.True(t, mgr.GetBitFlag(c))
	require.True(t, mgr.GetBitFlag(d))
	require.True(t, mgr.GetBitFlag(e))
}
