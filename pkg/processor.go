package pkg

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path"
	"slices"

	"github.com/ansel1/merry/v2"
	"github.com/go-logr/logr"
	"trading212-parser.kimi450.com/pkg/config"
	"trading212-parser.kimi450.com/pkg/logging"
	"trading212-parser.kimi450.com/pkg/trading212"
)

func Process(logBundleBaseDir, configFilePath, ticker string) {
	logBundleDir := path.Join(logBundleBaseDir,
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

	log.Info("runnig",
		"logBundleDir", logBundleDir,
		"configFilePath", configFilePath,
		"ticker", ticker)

	configData, err := config.ParseConfigFile(configFilePath)
	if err != nil {
		log.Error(err, "failed to parse config")
		os.Exit(1)
	}

	bookkeeper := trading212.NewBookkeeper()

	// sort files by year to ensure correct processing
	slices.SortFunc(configData.HistoryFiles, func(first, second config.HistoryFile) int {
		return cmp.Compare(first.Year, second.Year)
	})

	for _, historyFile := range configData.HistoryFiles {
		log.Info("processing file", "year", historyFile.Year, "path", historyFile.Path)

		profits, err := processHistoryFile(log, bookkeeper, historyFile, ticker)
		if err != nil {
			log.Error(err, "failed to process file",
				"year", historyFile.Year, "path", historyFile.Path)
			os.Exit(1)
		}
		log.Info("profits",
			"year", historyFile.Year,
			"value", profits,
		)
	}
}

func processHistoryFile(log logr.Logger, bookkeeper trading212.BookKeeper,
	historyFile config.HistoryFile, ticker string) (trading212.Profits, error) {
	file, err := os.Open(historyFile.Path)
	if err != nil {
		return trading212.Profits{}, merry.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// read csv values using csv.Reader
	csvReader := trading212.NewScanner(file)
	for csvReader.Scan() {
		record, err := csvReader.ToRecord()
		if err != nil {
			return trading212.Profits{}, merry.Errorf("failed to process file: %w", err)
		}

		if ticker != "" && ticker != record.Ticker {
			// if a ticker is passed, and if the record is not of the wanted ticker
			// skip the record
			continue
		}

		err = bookkeeper.FindOrCreateEntryAndProcess(log, record.Ticker, record)
		if err != nil {
			return trading212.Profits{}, err
		}
	}

	return bookkeeper.GetProfitForYear(historyFile.Year), nil
}
