package trading212

import (
	"github.com/ansel1/merry/v2"
	"github.com/go-logr/logr"
	"github.com/shopspring/decimal"
)

type Profits struct {
	Stock decimal.Decimal
	ETF   decimal.Decimal
}

type BookKeeperStruct struct {
	book map[string]PurchaseHistory
}

type BookKeeper interface {
	AddOrExtend(log logr.Logger, name string, purchaseHistory Record) error
	Print(log logr.Logger)
	GetProfitForYear(year int) Profits
}

func (b *BookKeeperStruct) Get(key string) PurchaseHistory {
	return b.book[key]
}

func NewBookkeeper() BookKeeper {
	return &BookKeeperStruct{book: make(map[string]PurchaseHistory)}
}

func (b *BookKeeperStruct) AddOrExtend(log logr.Logger, name string, record Record) error {
	_, ok := b.book[name]
	if !ok {
		b.book[name] = NewPurchaseHistory(NewRecordQueue())
	}
	purchaseHistory := b.book[name]

	err := purchaseHistory.Process(log, &record)
	if err != nil {
		return merry.Errorf("failed to update purchase history: %w", err)
	}
	return nil
}

func (b *BookKeeperStruct) Print(log logr.Logger) {

	for k, v := range b.book {
		for _, v2 := range v.GetRecordQueue().GetQueue() {
			log.Info("test", "k", k,
				"NoOfShares", v2.NoOfShares,
				"Price", v2.PriceShare,
				"Total", v2.Total,
				// "CurrencyConversionFee", v2.CurrencyConversionFee,
				// "ExchangeRate", v2.ExchangeRate,
				"profit", v.GetProfitForYear(2024),
			)

		}
	}

}

func (b *BookKeeperStruct) GetProfitForYear(year int) Profits {
	profits := Profits{
		ETF:   decimal.NewFromInt(0),
		Stock: decimal.NewFromInt(0),
	}
	for _, ph := range b.book {
		// get profit for each purchase history item for the year and sum it up
		yearlyProfit := ph.GetProfitForYear(year)
		profits.ETF = profits.ETF.Add(yearlyProfit.ETF)
		profits.Stock = profits.Stock.Add(yearlyProfit.Stock)
	}
	return profits
}
