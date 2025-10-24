package pkg

import (
	"cmp"
	"context"
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

type Report struct {
	ProfitsData          map[int]trading212.StockSummary
	SaleAggregatesData   map[int]trading212.StockSummary
	LossAggregatesData   map[int]trading212.StockSummary
	ProfitAggregatesData map[int]trading212.StockSummary
}

func getLog(logBundleBaseDir string) (logr.Logger, string, error) {
	logBundleDir := path.Join(logBundleBaseDir,
		config.GetDateTimePrefixForFile()+"-log-bundle")
	if _, err := os.Stat(logBundleDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(logBundleDir, os.ModePerm)
		if err != nil {
			return logr.Logger{}, "", fmt.Errorf("failed to create log bundle directory: %w", err)
		}
	}

	logFilePath := path.Join(logBundleDir,
		fmt.Sprintf("%s-script-logs.txt", config.GetDateTimePrefixForFile()))
	_, log, err := logging.GetDefaultFileAndConsoleLogger(logFilePath, false)
	if err != nil {
		return logr.Logger{}, "", fmt.Errorf("failed to setup logger: %w", err)
	}
	return log, logBundleDir, nil
}

func Process(logBundleBaseDir, configFilePath string, allowTickers, skipTickers []string) {
	var err error
	log := logr.FromContextOrDiscard(context.TODO())
	logBundleDir := ""
	if logBundleBaseDir != "" {
		log, logBundleDir, err = getLog(logBundleBaseDir)
		if err != nil {
			panic(err)
		}
	}

	log.Info("log bundle directory", "filePath", logBundleDir)

	log.Info("running",
		"logBundleDir", logBundleDir,
		"configFilePath", configFilePath,
		"allowTickers", allowTickers,
		"skipTickers", skipTickers)

	configData, err := config.ParseConfigFile(configFilePath)
	if err != nil {
		log.Error(err, "failed to parse config")
		os.Exit(1)
	}

	_ = processAllHistoryFiles(log, allowTickers, skipTickers, *configData)
}

func processAllHistoryFiles(log logr.Logger, allowTickers, skipTickers []string, configData config.Config) Report {
	summary := Report{
		ProfitsData:          make(map[int]trading212.StockSummary),
		SaleAggregatesData:   make(map[int]trading212.StockSummary),
		LossAggregatesData:   make(map[int]trading212.StockSummary),
		ProfitAggregatesData: make(map[int]trading212.StockSummary),
	}
	bookkeeper := trading212.NewBookkeeper()

	// sort files by year to ensure correct processing
	slices.SortFunc(configData.HistoryFiles, func(first, second config.HistoryFile) int {
		return cmp.Compare(first.Year, second.Year)
	})

	for _, historyFile := range configData.HistoryFiles {
		log.Info("processing file", "year", historyFile.Year, "path", historyFile.Path)

		saleAggregates, profitAggregates, lossAggregates, profits, err := processHistoryFile(log, bookkeeper, historyFile, allowTickers, skipTickers)
		if err != nil {
			log.Error(err, "failed to process file",
				"year", historyFile.Year, "path", historyFile.Path)
			os.Exit(1)
		}

		log.Info("summary",
			"year", historyFile.Year,
			"profits", profits,
		)

		log.Info("summary",
			"year", historyFile.Year,
			"sale aggregates", saleAggregates,
		)

		log.Info("summary",
			"year", historyFile.Year,
			"loss aggregates", lossAggregates,
		)

		log.Info("summary",
			"year", historyFile.Year,
			"profit aggregates", profitAggregates,
		)

		summary.ProfitsData[historyFile.Year] = profits
		summary.SaleAggregatesData[historyFile.Year] = saleAggregates
		summary.LossAggregatesData[historyFile.Year] = lossAggregates
		summary.ProfitAggregatesData[historyFile.Year] = profitAggregates
	}
	return summary
}

func processHistoryFile(log logr.Logger, bookkeeper trading212.BookKeeper,
	historyFile config.HistoryFile,
	allowTickers, skipTickers []string) (trading212.StockSummary,
	trading212.StockSummary,
	trading212.StockSummary,
	trading212.StockSummary, error) {

	file, err := os.Open(historyFile.Path)
	if err != nil {
		return trading212.StockSummary{}, trading212.StockSummary{},
			trading212.StockSummary{}, trading212.StockSummary{},
			merry.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// read csv values using csv.Reader
	csvReader := trading212.NewScanner(file)
	for csvReader.Scan() {
		record, err := csvReader.ToRecord()
		if err != nil {
			return trading212.StockSummary{}, trading212.StockSummary{},
				trading212.StockSummary{}, trading212.StockSummary{},
				merry.Errorf("failed to process file: %w", err)
		}

		if len(skipTickers) > 0 && valueInList(record.Ticker, skipTickers) {
			continue
		}

		if len(allowTickers) == 0 || valueInList(record.Ticker, allowTickers) {
			err = bookkeeper.FindOrCreateEntryAndProcess(log, record.Ticker, record)
			if err != nil {
				return trading212.StockSummary{}, trading212.StockSummary{},
					trading212.StockSummary{}, trading212.StockSummary{},
					err
			}
		}
	}

	return bookkeeper.GetSaleAggregatesForYear(historyFile.Year),
		bookkeeper.GetProfitAggregatesForYear(historyFile.Year),
		bookkeeper.GetLossAggregatesForYear(historyFile.Year),
		bookkeeper.GetProfitForYear(historyFile.Year),
		nil
}

func valueInList(value string, list []string) bool {
	for _, i := range list {
		if value == i {
			return true
		}
	}
	return false
}
