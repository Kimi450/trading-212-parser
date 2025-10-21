package pkg

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"trading212-parser.kimi450.com/pkg/config"
	"trading212-parser.kimi450.com/pkg/trading212"
)

func TestProcessHistoryFileLIFO(t *testing.T) {
	log := logr.FromContextOrDiscard(context.TODO())

	bookkeeper := trading212.NewBookkeeper()

	historyFile := config.HistoryFile{
		Year: 2024,
		Path: "../test-data/testdata-lifo-only.csv",
	}
	saleAggregates, lossAggregates, profits, err := processHistoryFile(log, bookkeeper, historyFile, "")

	assert.NoError(t, err)
	t.Log(profits.Overall)
	assertEqualDecimals(t, decimal.NewFromInt(39), profits.Overall)
	assertEqualDecimals(t, decimal.NewFromInt(56), saleAggregates.Overall)
	assertEqualDecimals(t, decimal.NewFromInt(0), lossAggregates.Overall)

}

func TestProcessHistoryFileFIFO(t *testing.T) {
	log := logr.FromContextOrDiscard(context.TODO())

	bookkeeper := trading212.NewBookkeeper()

	historyFile := config.HistoryFile{
		Year: 2024,
		Path: "../test-data/testdata-fifo-only.csv",
	}
	saleAggregates, lossAggregates, profits, err := processHistoryFile(log, bookkeeper, historyFile, "")

	assert.NoError(t, err)
	t.Log(profits.Overall)
	assertEqualDecimals(t, decimal.NewFromInt(42), profits.Overall)
	assertEqualDecimals(t, decimal.NewFromInt(56), saleAggregates.Overall)
	assertEqualDecimals(t, decimal.NewFromInt(0), lossAggregates.Overall)

}

func TestProcessAllHistoryFiles(t *testing.T) {
	log := logr.FromContextOrDiscard(context.TODO())

	configData := config.Config{
		HistoryFiles: []config.HistoryFile{
			{
				Year: 2025,
				Path: "../test-data/testdata-2025.csv",
			},
			{
				Year: 2022,
				Path: "../test-data/testdata-2022.csv",
			},
			{
				Year: 2023,
				Path: "../test-data/testdata-2023.csv",
			},
		},
	}

	resultMap := map[int]decimal.Decimal{
		2022: decimal.NewFromInt(-30),
		2023: decimal.NewFromInt(70),
		2025: decimal.NewFromInt(127),
	}

	saleAggregatesResultMap := map[int]decimal.Decimal{
		2022: decimal.NewFromInt(100),
		2023: decimal.NewFromInt(128),
		2025: decimal.NewFromInt(141),
	}

	lossAggregatesResultMap := map[int]decimal.Decimal{
		2022: decimal.NewFromInt(-65),
		2023: decimal.NewFromInt(0),
		2025: decimal.NewFromInt(-4),
	}

	summary := processAllHistoryFiles(log, "", configData)

	for _, historyFile := range configData.HistoryFiles {
		actualProfitValue := summary.ProfitsData[historyFile.Year].Overall
		expectedProfitValue := resultMap[historyFile.Year]

		t.Log("testing profit",
			"year", historyFile.Year,
			"expected", actualProfitValue,
			"actual", expectedProfitValue)
		assertEqualDecimals(t, expectedProfitValue, actualProfitValue)

		actualSaleAggregateValue := summary.SaleAggregatesData[historyFile.Year].Overall
		expectedSaleAggregateValue := saleAggregatesResultMap[historyFile.Year]

		t.Log("testing sale aggregate data",
			"year", historyFile.Year,
			"expected", actualSaleAggregateValue,
			"actual", expectedSaleAggregateValue)
		assertEqualDecimals(t, expectedSaleAggregateValue, actualSaleAggregateValue)

		actualLossAggregateValue := summary.LossAggregatesData[historyFile.Year].Overall
		expectedLossAggregateValue := lossAggregatesResultMap[historyFile.Year]

		t.Log("testing loss aggregate data",
			"year", historyFile.Year,
			"expected", actualLossAggregateValue,
			"actual", expectedLossAggregateValue)
		assertEqualDecimals(t, expectedLossAggregateValue, actualLossAggregateValue)
	}
}

func TestProcessHistoryFileWashSaleEasy(t *testing.T) {
	log := logr.FromContextOrDiscard(context.TODO())

	bookkeeper := trading212.NewBookkeeper()

	historyFile := config.HistoryFile{
		Year: 2024,
		Path: "../test-data/testdata-wash-sale.csv",
	}
	saleAggregates, lossAggregates, profits, err := processHistoryFile(log, bookkeeper, historyFile, "")

	assert.NoError(t, err)
	t.Log(profits.Overall)

	assertEqualDecimals(t, decimal.NewFromInt(100), profits.Overall)
	assertEqualDecimals(t, decimal.NewFromInt(1540), saleAggregates.Overall)
	assertEqualDecimals(t, decimal.NewFromInt(-100), lossAggregates.Overall)

}

// makes the error message more readable
func assertEqualDecimals(t *testing.T, expected, actual decimal.Decimal) {
	assert.Equal(t, expected.InexactFloat64(), actual.InexactFloat64())
}
