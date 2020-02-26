package datatypes

import "testing"

func TestISOCountryCode(t *testing.T) {
	var tests = []struct {
		value   string
		partial bool
		result  string
		correct bool
	}{
		{"", false, "", false},    // too short
		{"", true, "", true},      // too short, but partial --> ok
		{"_", false, "", false},   // wrong char class
		{"_", true, "", false},    // wrong char class
		{"N", false, "", false},   // too short
		{"N", true, "N", true},    // too short, but partial --> ok
		{"n", true, "N", true},    // lowercase gets converted
		{"NL", false, "NL", true}, // perfect
		{"NL", true, "NL", true},  // perfect
		{"nl", false, "NL", true}, // perfect & converted
		{"NLS", false, "", false}, // too long
		{"NLS", true, "", false},  // too long
	}

	for _, test := range tests {
		result, err := ISOCountryCode(test.value, test.partial)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("ISOCountryCode(%s, %t) expected %t, got %t", test.value, test.partial, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("ISOCountryCode(%s, %t) expected \"%s\", got \"%s\"", test.value, test.partial, test.result, result)
		}
	}
}

func TestISORegionCode(t *testing.T) {
	var tests = []struct {
		value   string
		partial bool
		result  string
		correct bool
	}{
		{"", false, "", false},            // too short
		{"", true, "", true},              // too short, but partial --> ok
		{"(", false, "", false},           // wrong char class
		{"(", true, "", false},            // wrong char class still
		{"N", false, "N", true},           // one letter will do
		{"n", false, "N", true},           // one letter will do, ucased for you
		{"9", false, "9", true},           // one digit will do
		{"-", false, "-", true},           // one connector will do
		{" AK ", false, "AK", true},       // Alaska works, spaces are killed
		{"US - AK", false, "US-AK", true}, // full story, no spaces
	}

	for _, test := range tests {
		result, err := ISORegionCode(test.value, test.partial)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("ISORegionCode(%s, %t) expected %t, got %t", test.value, test.partial, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("ISORegionCode(%s, %t) expected \"%s\", got \"%s\"", test.value, test.partial, test.result, result)
		}
	}
}

func TestICAOAirportCode(t *testing.T) {
	var tests = []struct {
		value   string
		partial bool
		result  string
		correct bool
	}{
		{"", false, "", false},          // too short
		{"", true, "", true},            // too short, but partial --> ok
		{"(", false, "", false},         // wrong char class
		{"(", true, "", false},          // wrong char class still
		{"N", true, "N", true},          // one letter will do
		{"N", false, "", false},         // right char class but too short
		{"n", true, "N", true},          // one letter will do, ucased for you
		{"9", true, "9", true},          // one digit will do
		{"-", true, "", false},          // one connector will do
		{" EHAM ", false, "EHAM", true}, // Amsterdam works, spaces are killed
		{"EHAM", false, "EHAM", true},   // full story, no spaces
	}

	for _, test := range tests {
		result, err := ICAOAirportCode(test.value, test.partial)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("ICAOAirportCode(%s, %t) expected %t, got %t", test.value, test.partial, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("ICAOAirportCode(%s, %t) expected \"%s\", got \"%s\"", test.value, test.partial, test.result, result)
		}
	}
}

func TestIATAAirportCode(t *testing.T) {
	var tests = []struct {
		value   string
		partial bool
		result  string
		correct bool
	}{
		{"", false, "", false},        // too short
		{"", true, "", true},          // too short, but partial --> ok
		{"(", false, "", false},       // wrong char class
		{"(", true, "", false},        // wrong char class still
		{"N", true, "N", true},        // one letter will do
		{"N", false, "", false},       // right char class but too short
		{"n", true, "N", true},        // one letter will do, ucased for you
		{"9", true, "", false},        // wrong char class
		{"-", true, "", false},        // wrong char class
		{" AMS ", false, "AMS", true}, // Amsterdam works, spaces are killed
		{"EHAM", false, "", false},    // too long
	}

	for _, test := range tests {
		result, err := IATAAirportCode(test.value, test.partial)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("IATAAirportCode(%s, %t) expected %t, got %t", test.value, test.partial, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("IATAAirportCode(%s, %t) expected \"%s\", got \"%s\"", test.value, test.partial, test.result, result)
		}
	}
}

func TestRunwayCode(t *testing.T) {
	var tests = []struct {
		value   string
		partial bool
		result  string
		correct bool
	}{
		{"", false, "", false},        // too short
		{"", true, "", true},          // too short, but partial --> ok
		{"(", false, "", false},       // wrong char class
		{"(", true, "", false},        // wrong char class still
		{"N", true, "N", true},        // one letter will do
		{"n", true, "N", true},        // one letter will do, ucased for you
		{"9", true, "9", true},        // one digit will do
		{"-", true, "-", true},        // one separator will do
		{"27L", false, "27L", true},   // 27 left works
		{" 27L ", false, "27L", true}, // 27 left works, spaces are killed
	}

	for _, test := range tests {
		result, err := RunwayCode(test.value, test.partial)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("RunCode(%s, %t) expected %t, got %t", test.value, test.partial, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("RunwayCode(%s, %t) expected \"%s\", got \"%s\"", test.value, test.partial, test.result, result)
		}
	}
}

func TestRunwayLength(t *testing.T) {
	var tests = []struct {
		value   string
		empty   bool
		result  int
		correct bool
	}{
		{"", false, 0, false},         // too short
		{"", true, 0, true},           // too short, but optional
		{"(", false, 0, false},        // wrong char class
		{"N", false, 0, false},        // wrong char class
		{"N", true, 0, false},         // Wrong doesn't count as empty
		{"n", false, 0, false},        // wrong char class
		{"9", false, 9, true},         // one digit will do
		{"-", false, 0, false},        // not a number
		{"-8000", false, 0, false},    // negative number
		{"8000", false, 8000, true},   // 8000ft works
		{" 8000 ", false, 8000, true}, // 8000ft works, spaces are killed
		{"100000", false, 0, false},   // 100000ft (> 30KM) is too big
	}

	for _, test := range tests {
		result, err := RunwayLength(test.value, test.empty)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("RunwayLength(%s) expected %t, got %t", test.value, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("RunwayLength(%s) expected %d, got %d", test.value, test.result, result)
		}
	}
}

func TestRunwayWidth(t *testing.T) {
	var tests = []struct {
		value   string
		empty   bool
		result  int
		correct bool
	}{
		{"", false, 0, false},       // too short
		{"", true, 0, true},         // too short, empty allowed
		{"(", false, 0, false},      // wrong char class
		{"N", false, 0, false},      // wrong char class
		{"N", true, 0, false},       // wrong char class doesn't count as empty
		{"n", false, 0, false},      // wrong char class
		{"9", false, 9, true},       // one digit will do
		{"-", false, 0, false},      // not a number
		{"-300", false, 0, false},   // negative number
		{"300", false, 300, true},   // 300ft (roughly 100M) works
		{" 300 ", false, 300, true}, // 300ft (roughly 100M) works, spaces are killed
		{"100000", false, 0, false}, // 100000ft (> 30KM) is too big
	}

	for _, test := range tests {
		result, err := RunwayWidth(test.value, test.empty)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("RunwayWidth(%s) expected %t, got %t", test.value, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("RunwayWidth(%s) expected %d, got %d", test.value, test.result, result)
		}
	}
}

func TestRunwayLighted(t *testing.T) {
	var tests = []struct {
		value   string
		empty   bool
		result  bool
		correct bool
	}{
		{"", false, false, false},   // Empty not allowed
		{"", true, false, true},     // too short, empty allowed
		{"(", false, false, false},  // wrong char class
		{"T", false, false, false},  // wrong char class
		{"T", true, false, false},   // wrong char class doesn't count as empty
		{"n", false, false, false},  // wrong char class
		{"9", false, false, false},  // wrong char class
		{"-", false, false, false},  // wrong char class
		{"0", false, false, true},   // '0' counts as false
		{" 0 ", false, false, true}, // '0' counts as false, spaces are killed
		{"1", false, true, true},    // '1' counts as true
	}

	for _, test := range tests {
		result, err := RunwayLighted(test.value, test.empty)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("RunwayLighted(%s) expected %t, got %t", test.value, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("RunwayLighted(%s) expected %t, got %t", test.value, test.result, result)
		}
	}
}

func TestRunwayClosed(t *testing.T) {
	var tests = []struct {
		value   string
		empty   bool
		result  bool
		correct bool
	}{
		{"", false, false, false},   // Empty not allowed
		{"", true, false, true},     // too short, empty allowed
		{"(", false, false, false},  // wrong char class
		{"T", false, false, false},  // wrong char class
		{"T", true, false, false},   // wrong char class doesn't count as empty
		{"n", false, false, false},  // wrong char class
		{"9", false, false, false},  // wrong char class
		{"-", false, false, false},  // wrong char class
		{"0", false, false, true},   // '0' counts as false
		{" 0 ", false, false, true}, // '0' counts as false, spaces are killed
		{"1", false, true, true},    // '1' counts as true
	}

	for _, test := range tests {
		result, err := RunwayClosed(test.value, test.empty)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("RunwayLighted(%s) expected %t, got %t", test.value, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("RunwayLighted(%s) expected %t, got %t", test.value, test.result, result)
		}
	}
}

func TestRunwayHeading(t *testing.T) {
	var tests = []struct {
		value   string
		empty   bool
		result  int
		correct bool
	}{
		{"", false, 0, false},     // too short
		{"", true, 0, true},       // too short, empty allowed
		{"(", false, 0, false},    // wrong char class
		{"N", false, 0, false},    // wrong char class
		{"N", true, 0, false},     // wrong char class doesn't count as empty
		{"n", false, 0, false},    // wrong char class
		{"9", false, 9, true},     // one digit will do
		{"-", false, 0, false},    // not a number
		{"-60", false, 0, false},  // negative heading
		{"60", false, 60, true},   // 60deg works
		{" 60 ", false, 60, true}, // 60deg works, spaces are killed
		{"361", false, 0, false},  // 361deg is too big
	}

	for _, test := range tests {
		result, err := RunwayHeading(test.value, test.empty)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("RunwayHeading(%s) expected %t, got %t", test.value, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("RunwayHeading(%s) expected %d, got %d", test.value, test.result, result)
		}
	}
}

func TestRunwayThreshold(t *testing.T) {
	var tests = []struct {
		value   string
		empty   bool
		result  int
		correct bool
	}{
		{"", false, 0, false},         // too short
		{"", true, 0, true},           // too short, but optional
		{"(", false, 0, false},        // wrong char class
		{"N", false, 0, false},        // wrong char class
		{"N", true, 0, false},         // Wrong doesn't count as empty
		{"n", false, 0, false},        // wrong char class
		{"9", false, 9, true},         // one digit will do
		{"-", false, 0, false},        // not a number
		{"-8000", false, 0, false},    // negative number
		{"8000", false, 8000, true},   // 8000ft works
		{" 8000 ", false, 8000, true}, // 8000ft works, spaces are killed
		{"100000", false, 0, false},   // 100000ft (> 30KM) is too big
	}

	for _, test := range tests {
		result, err := RunwayThreshold(test.value, test.empty)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("RunwayThreshold(%s) expected %t, got %t", test.value, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("RunwayThreshold(%s) expected %d, got %d", test.value, test.result, result)
		}
	}
}

func TestLatitude(t *testing.T) {
	var tests = []struct {
		value   string
		empty   bool
		result  float64
		correct bool
	}{
		{"", false, 0.0, false},       // too short
		{"", true, 0.0, true},         // too short, but optional
		{"(", false, 0.0, false},      // wrong char class
		{"N", false, 0.0, false},      // wrong char class
		{"N", true, 0.0, false},       // Wrong doesn't count as empty
		{"n", false, 0.0, false},      // wrong char class
		{"9", false, 9.0, true},       // one digit will do
		{"-", false, 0.0, false},      // not a number
		{"-90.0", false, -90.0, true}, // -90deg works
		{"-91.0", false, 0.0, false},  // -91deg is too small
		{"90", false, 90.0, true},     // 90deg works
		{" 90 ", false, 90.0, true},   // 90deg works, spaces are killed
		{"91", false, 0, false},       // 91deg is too big
	}

	for _, test := range tests {
		result, err := Latitude(test.value, test.empty)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("Latitude(%s) expected %t, got %t", test.value, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("Latitude(%s) expected %f, got %f", test.value, test.result, result)
		}
	}
}

func TestLongitude(t *testing.T) {
	var tests = []struct {
		value   string
		empty   bool
		result  float64
		correct bool
	}{
		{"", false, 0.0, false},         // too short
		{"", true, 0.0, true},           // too short, but optional
		{"(", false, 0.0, false},        // wrong char class
		{"N", false, 0.0, false},        // wrong char class
		{"N", true, 0.0, false},         // Wrong doesn't count as empty
		{"n", false, 0.0, false},        // wrong char class
		{"9", false, 9.0, true},         // one digit will do
		{"-", false, 0.0, false},        // not a number
		{"-180.0", false, -180.0, true}, // -180deg works
		{"-181.0", false, 0.0, false},   // -181deg is too small
		{"0.0", false, 0.0, true},       // 0deg works
		{"0", false, 0.0, true},         // 0deg works, decimals not needed
		{"180", false, 180.0, true},     // 180deg works
		{" 180 ", false, 180.0, true},   // 180deg works, spaces are killed
		{"181", false, 0, false},        // 181deg is too big
	}

	for _, test := range tests {
		result, err := Longitude(test.value, test.empty)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("Longitude(%s) expected %t, got %t", test.value, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("Longitude(%s) expected %f, got %f", test.value, test.result, result)
		}
	}
}

func TestElevation(t *testing.T) {
	var tests = []struct {
		value   string
		empty   bool
		result  int
		correct bool
	}{
		{"", false, 0, false},           // too short
		{"", true, 0, true},             // too short, but optional
		{"(", false, 0, false},          // wrong char class
		{"N", false, 0, false},          // wrong char class
		{"N", true, 0, false},           // Wrong doesn't count as empty
		{"n", false, 0, false},          // wrong char class
		{"9", false, 9, true},           // one digit will do
		{"-", false, 0, false},          // not a number
		{"-45000", false, -45000, true}, // -45000ft works (very deep)
		{"-45001", false, 0, false},     // -45001 is too deep
		{"0", false, 0, true},           // 0deg works
		{"0.0", false, 0, true},         // 0deg works, decimals are ok
		{"30000", false, 30000, true},   // 30000ft works (very high)
		{" 30000 ", false, 30000, true}, // 30000ft works, spaces are killed
		{"30001", false, 0, false},      // 30001ft is too high
	}

	for _, test := range tests {
		result, err := Elevation(test.value, test.empty)
		if (test.correct && err != nil) || (!test.correct && err == nil) {
			t.Errorf("Elevation(%s) expected %t, got %t", test.value, test.correct, (err == nil))
		}
		if test.result != result {
			t.Errorf("Elevation(%s) expected %d, got %d", test.value, test.result, result)
		}
	}
}
