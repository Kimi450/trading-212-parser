package trading212

import "reflect"

type RecordQueue interface {
	Append(rec *Record)
	Remove(index int) *Record
	RemoveItem(record *Record) *Record
	Peek(index int) *Record
	IsEmpty() bool
	Size() int
	GetQueue() []*Record
}

type RecordQueueStruct struct {
	data []*Record
}

func NewRecordQueue() RecordQueue {
	return &RecordQueueStruct{
		data: make([]*Record, 0),
	}
}

func (q *RecordQueueStruct) Append(rec *Record) {
	q.data = append(q.data, rec)
}

func (q *RecordQueueStruct) Remove(index int) *Record {
	removed := q.data[index]
	q.data = append(q.data[:index], q.data[index+1:]...)
	return removed
}

func (q *RecordQueueStruct) RemoveItem(record *Record) *Record {
	newData := []*Record{}
	for _, oldRecord := range q.data {
		if reflect.DeepEqual(record, oldRecord) {
			continue
		}
		newData = append(newData, oldRecord)
	}

	q.data = newData
	return record
}

func (q *RecordQueueStruct) Peek(index int) *Record {
	return q.data[index]
}

func (q *RecordQueueStruct) IsEmpty() bool {
	return len(q.data) == 0
}

func (q *RecordQueueStruct) Size() int {
	return len(q.data)
}

func (q *RecordQueueStruct) GetQueue() []*Record {
	return q.data
}
