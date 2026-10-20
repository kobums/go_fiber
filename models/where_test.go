package models

import (
	"reflect"
	"testing"
)

func TestWhereClause(t *testing.T) {
	tests := []struct {
		item   Where
		query  string
		params []interface{}
	}{
		{Where{Column: "name", Value: "a", Compare: "="}, " and u_name = ?", []interface{}{"a"}},
		{Where{Column: "email", Value: "kw", Compare: "like"}, " and u_email like ?", []interface{}{"%kw%"}},
		{Where{Column: "date", Value: [2]string{"a", "b"}, Compare: "between"}, " and u_date between ? and ?", []interface{}{"a", "b"}},
		{Where{Column: "id", Value: []int64{1, 2}, Compare: "in"}, " and u_id in (?, ?)", []interface{}{int64(1), int64(2)}},
		{Where{Column: "id", Value: []string{"1) or (1=1"}, Compare: "in"}, " and u_id in (?)", []interface{}{"1) or (1=1"}},
		{Where{Column: "id", Value: []int{}, Compare: "in"}, " and 1=0", nil},
		{Where{Column: "id = 1 or 1", Value: 1, Compare: "="}, " and 1=0", nil},
		{Where{Column: "id", Value: 1, Compare: "= 1 or 1 ="}, " and 1=0", nil},
	}
	for _, tt := range tests {
		q, p := WhereClause("u_", tt.item)
		if q != tt.query || !reflect.DeepEqual(p, tt.params) {
			t.Errorf("WhereClause(%+v) = %q %v, want %q %v", tt.item, q, p, tt.query, tt.params)
		}
	}
}

func TestOrdering(t *testing.T) {
	if got := Ordering("id desc").Order; got != "id desc" {
		t.Errorf("Ordering(id desc) = %q", got)
	}
	if got := Ordering("id, (select 1)").Order; got != "" {
		t.Errorf("Ordering(injection) = %q", got)
	}
}
