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

func TestProcessLIFO(t *testing.T) {
	log := logr.FromContextOrDiscard(context.TODO())

	bookkeeper := trading212.NewBookkeeper()

	historyFile := config.HistoryFile{
		Year: 2024,
		Path: "../test-data/testdata-lifo-only.csv",
	}
	profits, err := processHistoryFile(log, bookkeeper, historyFile, "")

	assert.NoError(t, err)
	t.Log(profits.Overall)
	assert.True(t, profits.Overall.Equal(decimal.NewFromInt(39)))

}

func TestProcessFIFO(t *testing.T) {
	log := logr.FromContextOrDiscard(context.TODO())

	bookkeeper := trading212.NewBookkeeper()

	historyFile := config.HistoryFile{
		Year: 2024,
		Path: "../test-data/testdata-fifo-only.csv",
	}
	profits, err := processHistoryFile(log, bookkeeper, historyFile, "")

	assert.NoError(t, err)
	t.Log(profits.Overall)
	assert.True(t, profits.Overall.Equal(decimal.NewFromInt(42)))

}
