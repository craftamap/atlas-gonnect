package store

import "testing"

func TestNew(t *testing.T) {
	// ToDo: better tests
	testCases := []struct {
		DBType      string
		DatabaseUrl string
		expectError bool
	}{
		{
			DBType:      "sqlite3",
			DatabaseUrl: ":memory:",
			expectError: false,
		},
	}

	for _, testCase := range testCases {
		_, err := New(testCase.DBType, testCase.DatabaseUrl)
		if err != nil {
			t.Error(err)
		}
	}
}
