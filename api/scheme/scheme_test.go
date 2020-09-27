package scheme

import "testing"

const testGroup = "test-group"

var testVersionKinds = map[string][]string{
	"foo1": []string{
		"Bar1",
	},
	"foo2": []string{
		"Bar2",
	},
}

func TestIsKnownGroupVersion(t *testing.T) {
	testCases := []struct {
		name           string
		group          string
		version        string
		expectedResult bool
	}{
		{
			name:           "known group and version",
			group:          testGroup,
			version:        "foo1",
			expectedResult: true,
		},
		{
			name:           "known group, unknown version",
			group:          testGroup,
			version:        "unknown",
			expectedResult: false,
		},
		{
			name:           "unknown group and version",
			group:          "unknown",
			version:        "unknown",
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isKnownGroupVersion(testGroup, testVersionKinds, tc.group, tc.version)
			if result != tc.expectedResult {
				t.Errorf("expected result %v, got %v", tc.expectedResult, result)
			}
		})
	}
}

func TestIsKnownVersionKind(t *testing.T) {
	testCases := []struct {
		name           string
		version        string
		kind           string
		expectedResult bool
	}{
		{
			name:           "known version and kind",
			version:        "foo1",
			kind:           "Bar1",
			expectedResult: true,
		},
		{
			name:           "known version, unknown kind",
			version:        "foo1",
			kind:           "Unknown",
			expectedResult: false,
		},
		{
			name:           "unknown version and kind",
			version:        "unknown",
			kind:           "Unknown",
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isKnownVersionKind(testVersionKinds, tc.version, tc.kind)
			if result != tc.expectedResult {
				t.Errorf("expected result %v, got %v", tc.expectedResult, result)
			}
		})
	}
}

func TestParseAPIVersion(t *testing.T) {
	testCases := []struct {
		name            string
		input           string
		expectedGroup   string
		expectedVersion string
		expectedError   bool
	}{
		{
			name:          "empty input",
			expectedError: true,
		},
		{
			name:            "valid input",
			input:           "foo/bar",
			expectedGroup:   "foo",
			expectedVersion: "bar",
		},
		{
			name:          "empty version",
			input:         "foo/",
			expectedError: true,
		},
		{
			name:          "empty group",
			input:         "/bar",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group, version, err := ParseAPIVersion(tc.input)
			if (err != nil) != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err != nil)
			}
			if group != tc.expectedGroup || version != tc.expectedVersion {
				t.Errorf("expected group, version [%s, %s], got [%s, %s]",
					tc.expectedGroup, tc.expectedVersion,
					group, version,
				)
			}
		})
	}
}

func TestReadAPIVersionKindFromJSON(t *testing.T) {
	testCases := []struct {
		name               string
		input              string
		expectedAPIVersion string
		expectedKind       string
		expectedError      bool
	}{
		{
			name:          "empty input",
			expectedError: true,
		},
		{
			name:               "valid apiVersion and kind",
			input:              `{"apiVersion": "foo", "kind": "bar", "someField":"z"}`,
			expectedAPIVersion: "foo",
			expectedKind:       "bar",
		},
		{
			name:          "empty apiVersion",
			input:         `{"apiVersion": ""}`,
			expectedError: true,
		},
		{
			name:          "empty kind",
			input:         `{"apiVersion": "foo", "kind": ""}`,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apiVersion, kind, err := ReadAPIVersionKindFromJSON([]byte(tc.input))
			if (err != nil) != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err != nil)
			}
			if apiVersion != tc.expectedAPIVersion || kind != tc.expectedKind {
				t.Errorf("expected apiVersion, kind [%s, %s], got [%s, %s]",
					tc.expectedAPIVersion, tc.expectedKind,
					apiVersion, kind,
				)
			}
		})
	}
}
