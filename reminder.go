package main

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type Reminder struct {
	Who   string
	What  string
	Where string
	When  time.Time
	index int
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
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Reminder)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func serialize(pq PriorityQueue, file string) error {
	b, e1 := json.Marshal(pq)
	if e1 != nil {
		return e1
	}
	f, e2 := os.Create(file)
	if e2 != nil {
		return e2
	}
	_, e3 := f.Write(b)
	if e3 != nil {
		return e3
	}
	fmt.Printf("Wrote reminders to %v\n", f.Name())
	return nil
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

func initializeReminders(file string) (PriorityQueue, error) {
	pq, err := deserialize(file)
	if err != nil {
		pq = make(PriorityQueue, 0)
		if err = serialize(pq, file); err != nil {
			return pq, err
		}
	}
	heap.Init(&pq)
	return pq, nil
}

type Reminders struct {
	File     string
	Queue    PriorityQueue
	Activity chan *Reminder
}

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
	return rs.Queue[n-1].When, nil
}

func (rs *Reminders) Next() (*Reminder, error) {
	if _, e := rs.PeekNextTime(); e != nil {
		return nil, e
	}
	return heap.Pop(&(rs.Queue)).(*Reminder), nil
}

func (rs *Reminders) Add(who, what, where string, when time.Time) error {
	r := &Reminder{
		Who:   who,
		What:  what,
		Where: where,
		When:  when,
		index: 0,
	}
	heap.Push(&(rs.Queue), r)
	go func() { rs.Activity <- r }()
	return nil
}

func (rs *Reminders) Save() error {
	return serialize(rs.Queue, rs.File)
}

func (rs *Reminders) Watch() chan *Reminder {
	c := make(chan *Reminder)
	go func() {
		for {
			if rs.Queue.Len() > 0 {
				t, _ := rs.PeekNextTime()
				select {
				case <-time.After(t.Sub(time.Now())):
					r, _ := rs.Next()
					c <- r
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
	}()
	return c
}

func example() {
	r, e1 := InitializeReminders("reminders.json")
	if e1 != nil {
		log.Fatal(e1)
	}
	go func() {
		for reminder := range r.Watch() {
			fmt.Printf("%v\n", reminder)
			r.Save()
		}
	}()

	time.Sleep(2 * time.Second)
	r.Add("me again", "2 (2) 4 second task", "#mathematics", time.Now().Add(4*time.Second))
	time.Sleep(2 * time.Second)
	r.Add("me again", "1 (4) 1 second task", "#mathematics", time.Now().Add(1*time.Second))
	time.Sleep(2 * time.Second)
	r.Add("me again", "3 (6) 1 intrusive", "#mathematics", time.Now().Add(1*time.Second))

	time.Sleep(20 * time.Second)
}
