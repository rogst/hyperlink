package message

import "testing"

func TestNewKeyFromRandomLetters(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		name  string
		args  args
		check func(got string) bool
	}{
		{"Zero length", args{0}, func(got string) bool { return got == "" }},
		{"Ten characters", args{10}, func(got string) bool { return len(got) == 10 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewKeyFromRandomLetters(tt.args.length); !tt.check(got) {
				t.Errorf("NewKeyFromRandomLetters() [%s], got %s, ", tt.name, got)
			}
		})
	}
}
