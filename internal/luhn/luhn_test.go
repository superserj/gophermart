package luhn

import "testing"

func TestValid(t *testing.T) {
	valid := []string{"12345678903", "79927398713", "9278923470", "2377225624", "4561261212345467"}
	for _, n := range valid {
		if !Valid(n) {
			t.Errorf("%s must be valid", n)
		}
	}
	invalid := []string{"12345678902", "79927398714", "", "12a45", "abc"}
	for _, n := range invalid {
		if Valid(n) {
			t.Errorf("%s must be invalid", n)
		}
	}
}
