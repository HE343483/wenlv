package service

import "testing"

func TestPickOpenHours(t *testing.T) {
	cases := []struct {
		name   string
		bizExt map[string]any
		want   string
	}{
		{"优先 open_time", map[string]any{"open_time": "09:00-17:00", "opentime_today": "08:00-18:00"}, "09:00-17:00"},
		{"回退 opentime_today", map[string]any{"opentime_today": "08:00-18:00"}, "08:00-18:00"},
		{"空数组视为无值", map[string]any{"open_time": "[]"}, ""},
		{"全部缺失", map[string]any{"rating": "4.5"}, ""},
		{"nil 入参", nil, ""},
	}
	for _, c := range cases {
		if got := pickOpenHours(c.bizExt); got != c.want {
			t.Errorf("%s: pickOpenHours=%q, want %q", c.name, got, c.want)
		}
	}
}
