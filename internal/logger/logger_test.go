package logger

import "testing"

func TestInitialize(t *testing.T) {
	if err := Initialize("info"); err != nil {
		t.Fatalf("valid level: unexpected error %v", err)
	}
	if Log == nil {
		t.Fatal("Log must be set after Initialize")
	}
	if err := Initialize("nonsense"); err == nil {
		t.Fatal("invalid level must return error")
	}
}
