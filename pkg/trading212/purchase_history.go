package trading212

type PurchaseHistory interface {
	Total()
}

type PurchaseHistoryStruct struct {
	recordQueue RecordQueue
	extraData   string
}

func NewPurchaseHistory(recordQueue RecordQueue) PurchaseHistory {
	return PurchaseHistoryStruct{
		recordQueue: recordQueue,
	}
}

func (p PurchaseHistoryStruct) Total() {
}
