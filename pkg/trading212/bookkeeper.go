package trading212

type BookKeeperStruct struct {
	book map[string]PurchaseHistory
}

type BookKeeper interface {
	AddOrExtend(name string, purchaseHistory Record)
}

func (b *BookKeeperStruct) Get(key string) PurchaseHistory {
	return b.book[key]
}

func NewBookkeeper() BookKeeper {
	return &BookKeeperStruct{book: make(map[string]PurchaseHistory)}
}

func (b *BookKeeperStruct) AddOrExtend(name string, purchaseHistory Record) {
	if _, ok := b.book[name]; !ok {
		// b.book[name] = purchaseHistory
	}
}
