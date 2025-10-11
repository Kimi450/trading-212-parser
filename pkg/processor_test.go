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
	saleAggregates, profits, err := processHistoryFile(log, bookkeeper, historyFile, "")

	assert.NoError(t, err)
	t.Log(profits.Overall)
	assert.True(t, profits.Overall.Equal(decimal.NewFromInt(39)))
	assert.True(t, saleAggregates.Overall.Equal(decimal.NewFromInt(56)))

}

func TestProcessHistoryFileFIFO(t *testing.T) {
	log := logr.FromContextOrDiscard(context.TODO())

	bookkeeper := trading212.NewBookkeeper()

	historyFile := config.HistoryFile{
		Year: 2024,
		Path: "../test-data/testdata-fifo-only.csv",
	}
	saleAggregates, profits, err := processHistoryFile(log, bookkeeper, historyFile, "")

	assert.NoError(t, err)
	t.Log(profits.Overall)
	assert.True(t, profits.Overall.Equal(decimal.NewFromInt(42)))
	assert.True(t, saleAggregates.Overall.Equal(decimal.NewFromInt(56)))

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
		2022: decimal.NewFromInt(0),
		2023: decimal.NewFromInt(70),
		2025: decimal.NewFromInt(36),
	}

	saleAggregatesResultMap := map[int]decimal.Decimal{
		2022: decimal.NewFromInt(0),
		2023: decimal.NewFromInt(128),
		2025: decimal.NewFromInt(40),
	}

	summary := processAllHistoryFiles(log, "", configData)

	for _, historyFile := range configData.HistoryFiles {
		actualProfitValue := summary.ProfitsData[historyFile.Year].Overall
		expectedProfitValue := resultMap[historyFile.Year]

		t.Log("testing profit", "expected", actualProfitValue, "actual", expectedProfitValue)
		assert.True(t, expectedProfitValue.Equal(actualProfitValue))

		actualSaleAggregateValue := summary.SaleAggregatesData[historyFile.Year].Overall
		expectedSaleAggregateValue := saleAggregatesResultMap[historyFile.Year]

		t.Log("testing aggregate data", "expected", actualSaleAggregateValue, "actual", expectedSaleAggregateValue)
		assert.True(t, expectedSaleAggregateValue.Equal(actualSaleAggregateValue))
	}
}
