package cmd

import (
	"strings"
	"testing"

	"github.com/dims/lambdactl/api"
)

func TestFindInstanceByRefPrefersID(t *testing.T) {
	instances := []api.Instance{
		{ID: "id-1", Name: "trainer"},
		{ID: "trainer", Name: "other"},
	}

	inst, err := findInstanceByRef(instances, "trainer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst.ID != "trainer" {
		t.Fatalf("expected exact ID match, got %q", inst.ID)
	}
}

func TestFindInstanceByRefMatchesUniqueName(t *testing.T) {
	instances := []api.Instance{
		{ID: "id-1", Name: "trainer"},
	}

	inst, err := findInstanceByRef(instances, "trainer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst.ID != "id-1" {
		t.Fatalf("expected name match id-1, got %q", inst.ID)
	}
}

func TestFindInstanceByRefRejectsAmbiguousNames(t *testing.T) {
	instances := []api.Instance{
		{ID: "id-1", Name: "trainer"},
		{ID: "id-2", Name: "trainer"},
	}

	_, err := findInstanceByRef(instances, "trainer")
	if err == nil {
		t.Fatalf("expected ambiguous-name error")
	}
	if !strings.Contains(err.Error(), "id-1") || !strings.Contains(err.Error(), "id-2") {
		t.Fatalf("expected IDs in error, got %v", err)
	}
}
