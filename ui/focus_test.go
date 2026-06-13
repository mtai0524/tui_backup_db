package ui

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
)

func TestFocusStopsCollapsed(t *testing.T) {
	m := Model{dbType: "MySQL", advancedExpanded: false}
	got := m.focusStops()
	want := []int{0, 1, 2, 3, 4, stopAdvancedToggle, stopButton}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collapsed: got %v want %v", got, want)
	}
}

func TestFocusStopsExpandedNonSQLServer(t *testing.T) {
	m := Model{dbType: "MySQL", advancedExpanded: true}
	got := m.focusStops()
	want := []int{0, 1, 2, 3, 4, stopAdvancedToggle, 5, 6, 7, stopButton}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expanded non-mssql: got %v want %v", got, want)
	}
}

func TestFocusStopsExpandedSQLServer(t *testing.T) {
	m := Model{dbType: "SQL Server", advancedExpanded: true}
	got := m.focusStops()
	want := []int{0, 1, 2, 3, 4, stopAdvancedToggle, 5, 6, 7, 8, stopButton}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expanded mssql: got %v want %v", got, want)
	}
}

func TestShouldAutoExpand(t *testing.T) {
	mk := func(vals map[int]string) []textinput.Model {
		ins := make([]textinput.Model, 9)
		for i := range ins {
			ins[i] = textinput.New()
		}
		for i, v := range vals {
			ins[i].SetValue(v)
		}
		return ins
	}
	if shouldAutoExpand(mk(map[int]string{0: "localhost", 4: "db"})) {
		t.Fatal("only required fields set → should NOT auto-expand")
	}
	if !shouldAutoExpand(mk(map[int]string{7: "~/backups"})) {
		t.Fatal("optional field 7 set → should auto-expand")
	}
	if !shouldAutoExpand(mk(map[int]string{8: ".bak"})) {
		t.Fatal("optional field 8 set → should auto-expand")
	}
}
