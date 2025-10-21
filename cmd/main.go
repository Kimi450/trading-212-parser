package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ansel1/merry/v2"
	"trading212-parser.kimi450.com/pkg"
)

type arrayFlags []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (i *arrayFlags) String() string {
	return fmt.Sprint(*i)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (i *arrayFlags) Set(value string) error {
	// If we wanted to allow the flag to be set multiple times,
	// accumulating values, we would delete this if statement.
	// That would permit usages such as
	//	-deltaT 10s -deltaT 15s
	// and other combinations.
	if len(*i) > 0 {
		return errors.New("arrayFlags flag already set")
	}
	for _, dt := range strings.Split(value, ",") {
		*i = append(*i, dt)
	}
	return nil
}

type ScriptArgs struct {
	LogBundleBaseDir string
	Config           string
	AllowTickers     arrayFlags
	SkipTickers      arrayFlags
}

func (scriptArgs *ScriptArgs) parseArgs() error {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])

		flag.PrintDefaults()
	}
	flag.ErrHelp = errors.New("flag: help requested")

	cwd, err := os.Getwd()
	if err != nil {
		return merry.Errorf("failed to get working directory: %w", err)
	}

	var allowTickers arrayFlags
	flag.Var(&allowTickers, "allowTickers", "Process only the given ticker(s). Specify more as a comma separated list.")

	var skipTickers arrayFlags
	flag.Var(&skipTickers, "skipTickers", "Skip the given ticker(s). Specify more as a comma separated list.")

	logBundleBaseDir := flag.String("log-bundle-base-dir", cwd,
		"Base directory for the log bundle generated")

	config := flag.String("config",
		path.Join(cwd, "configs", "config.json"),
		"Location of the script's config")

	flag.Parse()

	scriptArgs.LogBundleBaseDir = *logBundleBaseDir
	scriptArgs.Config = *config
	scriptArgs.AllowTickers = allowTickers
	scriptArgs.SkipTickers = skipTickers

	return nil
}

func (scriptArgs *ScriptArgs) verifyExpectedFilesExist() error {
	filePaths := []string{
		scriptArgs.Config,
	}

	for _, filePath := range filePaths {
		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			return merry.Errorf("file does not exist '%s': %w", filePath, err)
		}
	}

	return nil
}

func main() {

	scriptArgs := &ScriptArgs{}
	err := scriptArgs.parseArgs()
	if err != nil {
		panic(fmt.Errorf("failed to parse args: %w", err))
	}

	err = scriptArgs.verifyExpectedFilesExist()
	if err != nil {
		panic(fmt.Errorf("failed to validate args: %w", err))
	}

	pkg.Process(scriptArgs.LogBundleBaseDir, scriptArgs.Config, scriptArgs.AllowTickers, scriptArgs.SkipTickers)
}
