package storage

import (
	"errors"
	"testing"
)

func TestAdmitRefusesAtOrBelowReserve(t *testing.T) {
	if err := Admit(Reserve + 1); err != nil {
		t.Fatal(err)
	}
	if err := Admit(Reserve); !errors.Is(err, ErrRefused) {
		t.Fatalf("got %v", err)
	}
	if err := Admit(Reserve - 1); !errors.Is(err, ErrRefused) {
		t.Fatalf("got %v", err)
	}
}
