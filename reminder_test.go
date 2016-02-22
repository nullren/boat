package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func randomFileName(name, suffix string) string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%s_%X.%s", name, b, suffix)
}

func TestReminderInsertionOrder(t *testing.T) {
	remindersFile := randomFileName("reminders", "json")
	os.Remove(remindersFile)

	rs, err := InitializeReminders(remindersFile)
	if err != nil {
		t.Error("Unexpected error: ", err)
	}

	// checking the order
	rs.Add("4", "4 seconds", "", time.Now().Add(4*time.Second))
	rs.Add("2", "2 seconds", "", time.Now().Add(2*time.Second))
	rs.Add("1", "1 seconds", "", time.Now().Add(1*time.Second))
	rs.Add("3", "3 seconds", "", time.Now().Add(3*time.Second))

	r, err := rs.Next()
	if err != nil || r.Who != "1" {
		t.Error("Unexpected first r: ", r, err)
	}

	r, err = rs.Next()
	if err != nil || r.Who != "2" {
		t.Error("Unexpected second r: ", r, err)
	}

	r, err = rs.Next()
	if err != nil || r.Who != "3" {
		t.Error("Unexpected third r: ", r, err)
	}

	r, err = rs.Next()
	if err != nil || r.Who != "4" {
		t.Error("Unexpected fourth r: ", r, err)
	}

	os.Remove(remindersFile)
}

func TestPeekTimeMatchesNextTime(t *testing.T) {
	remindersFile := randomFileName("reminders", "json")
	os.Remove(remindersFile)

	rs, err := InitializeReminders(remindersFile)
	if err != nil {
		t.Error("Unexpected error: ", err)
	}

	// add a reminder really far into the future
	rs.Add("300", "300 seconds", "", time.Now().Add(300*time.Second))
	rs.Add("1", "1 second", "", time.Now().Add(time.Second))
	rs.Save()

	peekTime, _ := rs.PeekNextTime()
	r, _ := rs.Next()

	if r.When != peekTime {
		t.Errorf("Next time does not match PeekNextTime: %v /= %v", r.When, peekTime)
	}

	os.Remove(remindersFile)
}

func TestNotifyInterruptsWatch(t *testing.T) {
	remindersFile := randomFileName("reminders", "json")
	os.Remove(remindersFile)

	rs, err := InitializeReminders(remindersFile)
	if err != nil {
		t.Error("Unexpected error: ", err)
	}

	var wg sync.WaitGroup

	go rs.Watch(func(r *Reminder) {
		defer wg.Done()
		// should get interrupted and the next reminder passed is 1
		if r.Who != "1" {
			t.Error("did not get the correct reminder")
		}
	})

	// add a reminder really far into the future
	rs.Notify(rs.Add("300", "300 seconds", "", time.Now().Add(300*time.Second)))

	wg.Add(1)
	rs.Notify(rs.Add("1", "1 second", "", time.Now().Add(time.Second)))

	timer := time.AfterFunc(5*time.Second, func() {
		t.Error("waited too long, probably no interrupt")
	})

	wg.Wait()
	timer.Stop()

	os.Remove(remindersFile)
}

func TestNotifyInterruptsWatchWithOldReminders(t *testing.T) {
	remindersFile := randomFileName("reminders", "json")
	os.Remove(remindersFile)

	rso, err := InitializeReminders(remindersFile)
	if err != nil {
		t.Error("Unexpected error: ", err)
	}

	// add a reminder really far into the future
	rso.Notify(rso.Add("300", "300 seconds", "", time.Now().Add(300*time.Second)))
	rso.Save()

	// start over

	rs, err := InitializeReminders(remindersFile)
	if err != nil {
		t.Error("Unexpected error: ", err)
	}

	var wg sync.WaitGroup

	go rs.Watch(func(r *Reminder) {
		defer wg.Done()
		// should get interrupted and the next reminder passed is 1
		if r.Who != "1" {
			t.Error("did not get the correct reminder")
		}
	})

	wg.Add(1)
	rs.Notify(rs.Add("1", "1 second", "", time.Now().Add(time.Second)))
	rs.Save()

	timer := time.AfterFunc(5*time.Second, func() {
		t.Error("waited too long, probably no interrupt")
	})

	wg.Wait()
	timer.Stop()

	os.Remove(remindersFile)
}
