package trading212

import (
	"github.com/ansel1/merry/v2"
	"github.com/go-logr/logr"
)

type BookKeeperStruct struct {
	book map[string]PurchaseHistory
}

type BookKeeper interface {
	AddOrExtend(name string, purchaseHistory Record) error
	Print(log logr.Logger)
}

func (b *BookKeeperStruct) Get(key string) PurchaseHistory {
	return b.book[key]
}

func NewBookkeeper() BookKeeper {
	return &BookKeeperStruct{book: make(map[string]PurchaseHistory)}
}

func (b *BookKeeperStruct) AddOrExtend(name string, record Record) error {
	_, ok := b.book[name]
	if !ok {
		b.book[name] = NewPurchaseHistory(NewRecordQueue())
	}
	purchaseHistory := b.book[name]

	err := purchaseHistory.Process(&record)
	if err != nil {
		return merry.Errorf("failed to update purchase history: %w", err)
	}
	return nil
}

func (b *BookKeeperStruct) Print(log logr.Logger) {

	for k, v := range b.book {
		for _, v2 := range v.GetRecordQueue().GetQueue() {
			log.Info("test", "k", k,
				"Total", v2.Total,
				"NoOfShares", v2.NoOfShares,
				"CurrencyConversionFee", v2.CurrencyConversionFee,
				"ExchangeRate", v2.ExchangeRate,
			)

		}
	}

}
