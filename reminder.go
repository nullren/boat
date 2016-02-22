package main

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Reminder struct {
	Who   string
	What  string
	Where string
	When  time.Time
}

type Reminders struct {
	File     string
	Queue    PriorityQueue
	Activity chan *Reminder
}

// From: https://golang.org/pkg/container/heap/
// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Reminder

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the earliest Reminder, so just check that
	// i comes before j.
	return pq[i].When.Before(pq[j].When)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Reminder)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// serialization methods

func serialize(pq PriorityQueue, file string) (int, error) {
	b, err := json.Marshal(pq)
	if err != nil {
		return -1, err
	}
	f, err := os.Create(file)
	if err != nil {
		return -1, err
	}
	w, err := f.Write(b)
	if err != nil {
		return w, err
	}
	return w, nil
}

func loadJson(file string, target interface{}) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(target)
}

func deserialize(file string) (PriorityQueue, error) {
	var pq PriorityQueue
	err := loadJson(file, &pq)
	return pq, err
}

// Initialization: deserialize reminders from disk if exists, and then
// re-heapify them in the queue.

func initializeReminders(file string) (PriorityQueue, error) {
	pq, err := deserialize(file)
	if err != nil {
		pq = make(PriorityQueue, 0)
		if _, err = serialize(pq, file); err != nil {
			return pq, err
		}
	}
	heap.Init(&pq)
	return pq, nil
}

// these methods below actually constitute our API

func InitializeReminders(file string) (*Reminders, error) {
	pq, err := initializeReminders(file)
	if err != nil {
		return nil, err
	}
	return &Reminders{
		File:     file,
		Queue:    pq,
		Activity: make(chan *Reminder),
	}, nil
}

func (rs *Reminders) PeekNextTime() (time.Time, error) {
	n := len(rs.Queue)
	if n < 1 {
		return time.Now(), fmt.Errorf("Nothing to peek")
	}
	return rs.Queue[0].When, nil
}

func (rs *Reminders) Next() (*Reminder, error) {
	if _, e := rs.PeekNextTime(); e != nil {
		return nil, e
	}
	return heap.Pop(&(rs.Queue)).(*Reminder), nil
}

func (rs *Reminders) Notify(r *Reminder) {
	go func() { rs.Activity <- r }()
}

func (rs *Reminders) Add(who, what, where string, when time.Time) *Reminder {
	r := &Reminder{
		Who:   who,
		What:  what,
		Where: where,
		When:  when,
	}
	heap.Push(&(rs.Queue), r)
	return r
}

func (rs *Reminders) Save() error {
	_, err := serialize(rs.Queue, rs.File)
	return err
}

func (rs *Reminders) Watch(handler func(*Reminder)) {
	for {
		if t, err := rs.PeekNextTime(); err == nil {
			select {
			case <-time.After(t.Sub(time.Now())):
				if r, err := rs.Next(); err == nil {
					go handler(r)
				}
				break
			case <-rs.Activity:
				break
			}
		} else {
			select {
			case <-rs.Activity:
				break
			}
		}
	}
}
