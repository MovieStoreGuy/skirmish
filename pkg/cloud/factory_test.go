package cloud_test

import (
	"testing"

	"github.com/MovieStoreGuy/skirmish/pkg/cloud"
	"github.com/MovieStoreGuy/skirmish/pkg/minions"
)

func TestInsertion(t *testing.T) {
	fact := make(cloud.Factory, 0)
	err := fact.RegisterMinion("example", func() minions.Minion {
		return &minions.Mock{}
	})
	if err != nil {
		t.Fatal("Failed to insert minion", err)
	}
}

func TestCreation(t *testing.T) {
	fact := make(cloud.Factory, 0)
	err := fact.RegisterMinion("example", func() minions.Minion {
		return &minions.Mock{}
	})
	if err != nil {
		t.Fatal("Failed to insert minion", err)
	}
	example, err := fact.CreateMinion("example")
	if err != nil {
		t.Fatal(err)
	}
	if example == nil {
		t.Error("example was return as nil")
	}
	if _, err = fact.CreateMinion("doesnotexist"); err == nil {
		t.Fatal("Should not have returned a result")
	}

}
