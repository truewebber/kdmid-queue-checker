package daemon

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetCronRules(t *testing.T) {
	t.Parallel()

	type testCase struct {
		startCheckFrom       string
		startCheckTil        string
		amountOfTriggersADay int
		expectedOutput       []string
		expectError          bool
	}

	testCases := []testCase{
		// Positive test cases
		{"08:00", "18:00", 5, []string{"0 0 8 * * *", "0 30 10 * * *", "0 0 13 * * *", "0 30 15 * * *", "0 0 18 * * *"}, false},
		{"09:00", "09:01", 1, []string{"0 0 9 * * *"}, false},
		{"12:00", "12:00", 1, []string{"0 0 12 * * *"}, false},
		{"00:00", "23:59", 24, []string{"0 0 0 * * *", "0 2 1 * * *", "0 4 2 * * *", "0 6 3 * * *", "0 8 4 * * *", "0 10 5 * * *", "0 12 6 * * *", "0 14 7 * * *", "0 16 8 * * *", "0 18 9 * * *", "0 20 10 * * *", "0 22 11 * * *", "0 24 12 * * *", "0 26 13 * * *", "0 28 14 * * *", "0 30 15 * * *", "0 32 16 * * *", "0 34 17 * * *", "0 36 18 * * *", "0 38 19 * * *", "0 40 20 * * *", "0 42 21 * * *", "0 44 22 * * *", "0 46 23 * * *"}, false},
		{"00:00", "00:00", 24, []string{"0 0 0 * * *", "0 2 1 * * *", "0 4 2 * * *", "0 6 3 * * *", "0 8 4 * * *", "0 10 5 * * *", "0 12 6 * * *", "0 14 7 * * *", "0 16 8 * * *", "0 18 9 * * *", "0 20 10 * * *", "0 22 11 * * *", "0 24 12 * * *", "0 26 13 * * *", "0 28 14 * * *", "0 30 15 * * *", "0 32 16 * * *", "0 34 17 * * *", "0 36 18 * * *", "0 38 19 * * *", "0 40 20 * * *", "0 42 21 * * *", "0 44 22 * * *", "0 46 23 * * *"}, false},
		{"00:00", "00:00", 25, []string{"0 0 0 * * *", "0 0 1 * * *", "0 0 2 * * *", "0 0 3 * * *", "0 0 4 * * *", "0 0 5 * * *", "0 0 6 * * *", "0 0 7 * * *", "0 0 8 * * *", "0 0 9 * * *", "0 0 10 * * *", "0 0 11 * * *", "0 0 12 * * *", "0 0 13 * * *", "0 0 14 * * *", "0 0 15 * * *", "0 0 16 * * *", "0 0 17 * * *", "0 0 18 * * *", "0 0 19 * * *", "0 0 20 * * *", "0 0 21 * * *", "0 0 22 * * *", "0 0 23 * * *", "0 0 0 * * *"}, false},
		{"00:00", "00:00", 1, []string{"0 0 0 * * *"}, false},
		{"10:00", "10:01", 2, []string{"0 0 10 * * *", "0 1 10 * * *"}, false},
		{"08:00", "09:00", 3, []string{"0 0 8 * * *", "0 30 8 * * *", "0 0 9 * * *"}, false},
		{"10:00", "11:00", 4, []string{"0 0 10 * * *", "0 20 10 * * *", "0 40 10 * * *", "0 0 11 * * *"}, false},
		{"14:00", "14:02", 3, []string{"0 0 14 * * *", "0 1 14 * * *", "0 2 14 * * *"}, false},
		{"00:00", "12:00", 6, []string{"0 0 0 * * *", "0 24 2 * * *", "0 48 4 * * *", "0 12 7 * * *", "0 36 9 * * *", "0 0 12 * * *"}, false},
		{"18:00", "08:00", 5, []string{"0 0 18 * * *", "0 30 21 * * *", "0 0 1 * * *", "0 30 4 * * *", "0 0 8 * * *"}, false},

		// Negative test cases
		{"0800", "1800", 5, nil, true},
		{"08:xx", "18:yy", 5, nil, true},
		{"08:00", "18:00", -1, nil, true},
		{"08:00", "18:00", 0, nil, true},
		{"25:00", "26:00", 5, nil, true},
		{"08:00", "09:00", 1000, nil, true},
		{"", "", 5, nil, true},
		{" 08:00", "18:00 ", 5, nil, true},
		{"10:00", "10:01", 3, nil, true},
	}

	for _, tt := range testCases {
		tt := tt

		t.Run(fmt.Sprintf("%s_%s_%d", tt.startCheckFrom, tt.startCheckTil, tt.amountOfTriggersADay), func(t *testing.T) {
			t.Parallel()

			c := &CheckSlot{}

			cronRules, err := c.getCronRules(tt.startCheckFrom, tt.startCheckTil, tt.amountOfTriggersADay)

			if (err != nil) != tt.expectError {
				t.Fatalf("expected error: %v, got: %v", tt.expectError, err)
			}

			if !tt.expectError && !reflect.DeepEqual(cronRules, tt.expectedOutput) {
				t.Errorf("expected: %v, got: %v", tt.expectedOutput, cronRules)
			}
		})
	}
}
