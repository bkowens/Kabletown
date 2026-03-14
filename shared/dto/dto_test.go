package dto

import (
	"encoding/json"
	"testing"
)

func TestUserDataDto_JSON(t *testing.T) {
	ud := UserDataDto{Rating: 5.5, Played: true}
	data, err := json.Marshal(ud)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("empty JSON")
	}
}

func TestImageTags_JSON(t *testing.T) {
	it := ImageTags{"Primary": "tag123"}
	data, err := json.Marshal(it)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]string
	json.Unmarshal(data, &out)
	if out["Primary"] != "tag123" {
		t.Error("tag mismatch")
	}
}

func TestNewPagedResult(t *testing.T) {
	pr := NewPagedResult([]string{"a"}, 10, 0, 20)
	if len(pr.Items) != 1 {
		t.Error("items len")
	}
}

func TestItemTypeValueByName(t *testing.T) {
	v, ok := ItemTypeValueByName["Movie"]
	if !ok || v != ItemTypeValueMovie {
		t.Error("Movie lookup")
	}
}

func TestNewPagedResult_NegativeTotal(t *testing.T) {
	pr := NewPagedResult([]string{"a"}, -5, 0, 20)
	if pr.TotalRecordCount != 0 {
		t.Error("negative total should be 0")
	}
}

func TestPagedResultSliceRange(t *testing.T) {
	pr := NewPagedResult([]string{"a", "b"}, 100, 10, 20)
	rangeStr := pr.SliceRange()
	if rangeStr == "" {
		t.Error("slice range empty")
	}
}
