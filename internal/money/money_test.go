package money

import (
	"encoding/json"
	"math"
	"testing"
)

func TestMarshalJSON(t *testing.T) {
	cases := map[Points]string{72998: "729.98", 50000: "500.00", 0: "0.00", 1: "0.01"}
	for p, want := range cases {
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("marshal %d: %v", p, err)
		}
		if string(b) != want {
			t.Fatalf("marshal %d = %s, want %s", p, b, want)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	cases := map[string]Points{"729.98": 72998, "500": 50000, "0.01": 1, "729.975": 72998}
	for in, want := range cases {
		var p Points
		if err := json.Unmarshal([]byte(in), &p); err != nil {
			t.Fatalf("unmarshal %s: %v", in, err)
		}
		if p != want {
			t.Fatalf("unmarshal %s = %d, want %d", in, p, want)
		}
	}
}

func TestFromFloatAndFloat(t *testing.T) {
	if FromFloat(729.98) != Points(72998) {
		t.Fatalf("FromFloat(729.98) = %d", FromFloat(729.98))
	}
	if FromFloat(500) != Points(50000) {
		t.Fatalf("FromFloat(500) = %d", FromFloat(500))
	}
	if float32(FromFloat(729.98).Float()) != float32(729.98) {
		t.Fatal("Float must round-trip to float32(729.98)")
	}
}

func TestFromFloatNaNInf(t *testing.T) {
	if FromFloat(math.NaN()) != Points(0) {
		t.Fatalf("FromFloat(NaN) = %d, want 0", FromFloat(math.NaN()))
	}
	if FromFloat(math.Inf(1)) != Points(0) {
		t.Fatalf("FromFloat(+Inf) = %d, want 0", FromFloat(math.Inf(1)))
	}
	if FromFloat(math.Inf(-1)) != Points(0) {
		t.Fatalf("FromFloat(-Inf) = %d, want 0", FromFloat(math.Inf(-1)))
	}
	if FromFloat(729.98) != Points(72998) {
		t.Fatalf("FromFloat(729.98) = %d, want 72998", FromFloat(729.98))
	}
}

func TestRoundTrip(t *testing.T) {
	orig := Points(72998)
	b, _ := json.Marshal(orig)
	var got Points
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got != orig {
		t.Fatalf("round-trip: %d != %d", got, orig)
	}
}
