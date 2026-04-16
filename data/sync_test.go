package data

import (
	"testing"

	"github.com/Kevin-Aguirre/powerlifting-api/model"
)

func TestDataStore_SetAndGet(t *testing.T) {
	ds := NewDataStore()

	// Initially nil
	if got := ds.DB(); got != nil {
		t.Error("new DataStore should return nil DB")
	}

	// Set and retrieve
	db := &Database{
		LifterHistory:   map[string]*model.Lifter{"Test": {Name: "Test"}},
		FederationMeets: make(map[string][]*model.Meet),
		MeetResults:     make(map[string][]*model.LifterMeetResult),
	}
	ds.set(db)

	got := ds.DB()
	if got == nil {
		t.Fatal("DB() returned nil after set")
	}
	if _, ok := got.LifterHistory["Test"]; !ok {
		t.Error("DB should contain Test lifter")
	}
}

func TestDataStore_AtomicSwap(t *testing.T) {
	ds := NewDataStore()

	db1 := &Database{
		LifterHistory:   map[string]*model.Lifter{"First": {Name: "First"}},
		FederationMeets: make(map[string][]*model.Meet),
		MeetResults:     make(map[string][]*model.LifterMeetResult),
	}
	db2 := &Database{
		LifterHistory:   map[string]*model.Lifter{"Second": {Name: "Second"}},
		FederationMeets: make(map[string][]*model.Meet),
		MeetResults:     make(map[string][]*model.LifterMeetResult),
	}

	ds.set(db1)
	if _, ok := ds.DB().LifterHistory["First"]; !ok {
		t.Error("should contain First")
	}

	ds.set(db2)
	if _, ok := ds.DB().LifterHistory["Second"]; !ok {
		t.Error("should contain Second after swap")
	}
	if _, ok := ds.DB().LifterHistory["First"]; ok {
		t.Error("should not contain First after swap")
	}
}
