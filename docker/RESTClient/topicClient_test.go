package main

import "testing"

func Test_getData(t *testing.T) {
	tests := []struct {
		name  string
		want  int
		want1 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getData()
			if got != tt.want {
				t.Errorf("getData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
