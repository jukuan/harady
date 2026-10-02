package util

import (
	"reflect"
	"testing"
)

func TestSplitList(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"a", []string{"a"}},
		{" a ; b ; c ", []string{"a", "b", "c"}},
		{";;", nil},
		{"a;", []string{"a"}},
		{";a", []string{"a"}},
		{"  ; ", nil},
		{"Мінск; Гомель", []string{"Мінск", "Гомель"}},
	}
	for _, c := range cases {
		got := SplitList(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("SplitList(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}
