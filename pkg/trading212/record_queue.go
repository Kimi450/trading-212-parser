package trading212

type RecordQueue interface {
	Enqueue(rec *Record)
	Dequeue() *Record
	Peak() *Record
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

func (q *RecordQueueStruct) Enqueue(rec *Record) {
	q.data = append(q.data, rec)
}

func (q *RecordQueueStruct) Dequeue() *Record {
	dequeued := q.data[0]
	q.data = q.data[1:]
	return dequeued
}

func (q *RecordQueueStruct) Peak() *Record {
	return q.data[0]
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
