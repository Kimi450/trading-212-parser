package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/ansel1/merry/v2"
	"github.com/go-logr/logr"
	"trading212-parser.kimi450.com/pkg/config"
	"trading212-parser.kimi450.com/pkg/logging"
	"trading212-parser.kimi450.com/pkg/trading212"
)

type ScriptArgs struct {
	LogBundleBaseDir string
	Config           string
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

	logBundleBaseDir := flag.String("log-bundle-base-dir", cwd,
		"Base directory for the log bundle generated")

	config := flag.String("config",
		path.Join(cwd, "configs", "config.json"),
		"Location of the script's config")

	flag.Parse()

	scriptArgs.LogBundleBaseDir = *logBundleBaseDir
	scriptArgs.Config = *config

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

func (scriptArgs *ScriptArgs) run() {
	logBundleDir := path.Join(scriptArgs.LogBundleBaseDir,
		config.GetDateTimePrefixForFile()+"-log-bundle")
	if _, err := os.Stat(logBundleDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(logBundleDir, os.ModePerm)
		if err != nil {
			panic(fmt.Errorf("failed to create log bundle directory: %w", err))
		}
	}

	logFilePath := path.Join(logBundleDir,
		fmt.Sprintf("%s-script-logs.txt", config.GetDateTimePrefixForFile()))
	_, log, err := logging.GetDefaultFileAndConsoleLogger(logFilePath, false)
	if err != nil {
		panic(fmt.Errorf("failed to setup logger: %w", err))
	}

	log.Info("log bundle directory", "filePath", logBundleDir)

	configData, err := config.ParseConfigFile(scriptArgs.Config)
	if err != nil {
		log.Error(err, "failed to parse config")
		os.Exit(1)
	}

	log.Info("script args passed", "scriptArgs", scriptArgs)

	for _, historyFile := range configData.HistoryFiles {
		log.Info("processing file", "year", historyFile.Year, "path", historyFile.Path)

		err := processFile(log, historyFile)
		if err != nil {
			log.Error(err, "failed to process file",
				"year", historyFile.Year, "path", historyFile.Path)
			os.Exit(1)
		}
	}
}

func processFile(log logr.Logger, historyFile config.HistoryFile) error {
	file, err := os.Open(historyFile.Path)
	if err != nil {
		return merry.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// read csv values using csv.Reader
	csvReader := trading212.NewScanner(file)
	for csvReader.Scan() {
		record, err := csvReader.ToRecord()
		if err != nil {
			return merry.Errorf("failed to process file: %w", err)
		}
		if record.StampDutyReserveTax != "" {
			log.Info("record", "data", record)
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

	scriptArgs.run()
}
