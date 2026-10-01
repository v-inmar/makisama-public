package strings_utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"testing"
)

func TestIsValidLength(t *testing.T) {
	tests := []struct {
		name     string // Description of the scenario
		input    string // The string to test
		min      int    // Minimum constraint
		max      int    // Maximum constraint
		expected bool   // What the function should return
	}{
		{
			name:     "Standard valid string within bounds",
			input:    "salamagi",
			min:      3,
			max:      10,
			expected: true,
		},
		{
			name:     "Exact minimum boundary limit",
			input:    "abc",
			min:      3,
			max:      5,
			expected: true,
		},
		{
			name:     "Exact maximum boundary limit",
			input:    "abcde",
			min:      2,
			max:      5,
			expected: true,
		},
		{
			name:     "String is too short",
			input:    "yo",
			min:      3,
			max:      10,
			expected: false,
		},
		{
			name:     "String is too long",
			input:    "thisstringiswaytoolong",
			min:      1,
			max:      5,
			expected: false,
		},
		{
			name:     "Valid space padding gets trimmed down cleanly",
			input:    "   hello   ",
			min:      5,
			max:      5,
			expected: true, // "hello" is exactly 5 runes
		},
		{
			name:     "Empty spaces collapse to zero length and fail min bound",
			input:    "      ",
			min:      1,
			max:      5,
			expected: false, // Becomes "" which has a length of 0
		},
		{
			name:     "Multi-byte Unicode / Emoji validation correctness",
			input:    "🚀🚀🚀",
			min:      3,
			max:      3,
			expected: true, // 3 runes! (A regular byte check would erroneously evaluate this as 12 bytes)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsValidLength(tt.input, tt.min, tt.max)

			if actual != tt.expected {
				t.Errorf(
					"IsValidLength(%q, min: %d, max: %d) failed: expected %v, got %v",
					tt.input, tt.min, tt.max, tt.expected, actual,
				)
			}
		})
	}

}

func TestCapitalize(t *testing.T) {
	// 1. Establish the table of diverse scenarios
	tests := []struct {
		name     string // Contextual label for the test case
		input    string // String passed into Capitalize()
		expected string // The expected output string
	}{
		{
			name:     "Standard lowercase sentence",
			input:    "hello world from go",
			expected: "Hello World From Go",
		},
		{
			name:     "Already capitalized words remain unchanged",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Mixed case words preserve their inner uppercase styling",
			input:    "hElLo wOrLd",
			expected: "HElLo WOrLd", // Validates that only index [0] is altered
		},
		{
			name:     "Single-letter words are successfully capitalized",
			input:    "a b c d",
			expected: "A B C D",
		},
		{
			name:     "Empty string returns an empty string safely",
			input:    "",
			expected: "",
		},
		{
			name:     "Strings with only whitespace resolve to empty strings",
			input:    "     ",
			expected: "",
		},
		{
			name:     "Irregular whitespace, tabs, and newlines are normalized to single spaces",
			input:    "hello   \t   world\nline",
			expected: "Hello World Line", // Verifies the behavior of strings.Fields
		},
		{
			name:     "Words starting with numbers or symbols are ignored but unharmed",
			input:    "123go !hello #world",
			expected: "123go !hello #world",
		},
		{
			name:     "Multi-byte Unicode / International characters capitalize correctly",
			input:    "çilek özgürlük şoför",
			expected: "Çilek Özgürlük Şoför", // Verifies rune-level unicode mapping works
		},
		{
			name:     "Words starting with an Emoji are unharmed",
			input:    "🚀rocket fire🔥 work",
			expected: "🚀rocket Fire🔥 Work", // Emojis don't have uppercase variants, handles safely
		},
	}

	// 2. Iterate through each test case scenario
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Capitalize(tt.input)

			// 3. Execute assertion
			if actual != tt.expected {
				t.Errorf("Capitalize(%q) failed:\n  expected: %q\n  got:      %q", tt.input, tt.expected, actual)
			}
		})
	}
}

func TestSubstring(t *testing.T) {
	// 1. Establish the table of inputs, bounds, and expected outputs
	tests := []struct {
		name     string // Contextual label for the test case
		input    string // String passed into Substring()
		max      int    // The maximum length constraint
		expected string // The expected truncated or trimmed output
	}{
		{
			name:     "String length is strictly less than max",
			input:    "golang",
			max:      10,
			expected: "golang",
		},
		{
			name:     "String length is exactly equal to max",
			input:    "safari",
			max:      6,
			expected: "safari",
		},
		{
			name:     "String length exceeds max (Standard truncation)",
			input:    "salamagi project",
			max:      8,
			expected: "salamagi",
		},
		{
			name:     "Empty string handling with positive max",
			input:    "",
			max:      5,
			expected: "",
		},
		{
			name:     "Whitespace-only string collapses to empty string",
			input:    "      ",
			max:      3,
			expected: "",
		},
		{
			name:     "Surrounding spaces are trimmed before calculating max length constraints",
			input:    "   hello world   ",
			max:      5,
			expected: "hello", // Trims to "hello world", then truncates down to 5 runes
		},
		{
			name:     "Multi-byte Unicode / Emojis truncate accurately by character count",
			input:    "👋🌲🚀🔥💎",
			max:      3,
			expected: "👋🌲🚀", // 3 runes! A byte-based slice would produce unreadable character corruption
		},
		{
			name:     "Max threshold set to zero returns empty string cleanly",
			input:    "intercept",
			max:      0,
			expected: "",
		},
		{
			name:     "Negative max returns empty string safely",
			input:    "intercept",
			max:      -1,
			expected: "",
		},
	}

	// 2. Iterate through the test collection matrix
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Substring(tt.input, tt.max)

			// 3. Execute assertion
			if actual != tt.expected {
				t.Errorf(
					"Substring(%q, %d) failed:\n  expected: %q\n  got:      %q",
					tt.input, tt.max, tt.expected, actual,
				)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	// 1. Establish the table of inputs and expected boolean outcomes
	tests := []struct {
		name     string // Contextual label for the test case
		input    string // The email string passed into IsValidEmail()
		expected bool   // The expected verification result
	}{
		{
			name:     "Standard valid lowercase email",
			input:    "test@example.com",
			expected: true,
		},
		{
			name:     "Mixed-case valid email is normalized and passes",
			input:    "User.Name+Filter@Example.Co.Uk",
			expected: true, // Verifies strings.ToLower behaves correctly
		},
		{
			name:     "Valid email surrounded by whitespace is cleaned and passes",
			input:    "   hello@gopher.dev   ",
			expected: true, // Verifies strings.TrimSpace behaves correctly
		},
		{
			name:     "Completely empty string fails immediately",
			input:    "",
			expected: false,
		},
		{
			name:     "Whitespace-only string collapses to empty and fails",
			input:    "       ",
			expected: false,
		},
		{
			name:     "Missing domain suffix fails",
			input:    "username@",
			expected: false,
		},
		{
			name:     "Missing commercial at symbol fails",
			input:    "username.example.com",
			expected: false,
		},
		{
			name:     "Multiple at symbols fail the parser",
			input:    "user@name@example.com",
			expected: false,
		},
		{
			name:     "The Name Display Trap is caught and rejected",
			input:    "Gopher <gopher@example.com>",
			expected: false, // 🎯 Verifies that your `addr.Address == email` guard works!
		},
		{
			name:     "IP address domains are technically valid under RFC 5322",
			input:    "admin@192.168.1.1",
			expected: true,
		},
	}

	// 2. Iterate through each test case scenario
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsValidEmail(tt.input)

			// 3. Execute assertion
			if actual != tt.expected {
				t.Errorf("IsValidEmail(%q) failed: expected %v, got %v", tt.input, tt.expected, actual)
			}
		})
	}
}

func TestObfuscateEmail(t *testing.T) {
	// 1. Set up the table matrix covering standard inputs, boundaries, and errors
	tests := []struct {
		name     string // Description of the test scenario
		input    string // The email input string
		expected string // The expected obfuscated string outcome
	}{
		{
			name:     "Standard long local part",
			input:    "salamagi@example.com",
			expected: "s******i@example.com",
		},
		{
			name:     "Three letter local part (minimum for middle masking)",
			input:    "bob@test.org",
			expected: "b*b@test.org",
		},
		{
			name:     "Two letter local part boundary (should become all stars)",
			input:    "go@golang.org",
			expected: "**@golang.org",
		},
		{
			name:     "One letter local part boundary (should become a single star)",
			input:    "i@apple.com",
			expected: "*@apple.com",
		},
		{
			name:     "Invalid email string containing no at symbol",
			input:    "plainstringwithoutanatsymbol",
			expected: "plainstringwithoutanatsymbol", // Returns unmodified
		},
		{
			name:     "Invalid email string containing multiple at symbols",
			input:    "user@name@domain.com",
			expected: "user@name@domain.com", // Returns unmodified
		},
		{
			name:     "Multi-byte Unicode / Emoji local part masking",
			input:    "👋🌲🚀🔥💎@galaxy.io",
			expected: "👋***💎@galaxy.io", // Preserves first and last rune perfectly
		},
		{
			name:     "Empty local part string",
			input:    "@emptyuser.com",
			expected: "@emptyuser.com", // 0 runes <= 2, repeats "*" 0 times
		},
	}

	// 2. Execute the test runner loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ObfuscateEmail(tt.input)

			// 3. Evaluate results
			if actual != tt.expected {
				t.Errorf(
					"ObfuscateEmail(%q) failed:\n  expected: %q\n  got:      %q",
					tt.input, tt.expected, actual,
				)
			}
		})
	}
}

func TestNormalizeWebsiteURL(t *testing.T) {
	// 1. Define the table containing inputs, expected string modifications, and error messages
	tests := []struct {
		name        string // Description of the test scenario
		input       string // Raw input string passed to the function
		expectedURL string // The expected clean URL string outcome
		expectedErr string // The exact error message expected, or "" if no error
	}{
		{
			name:        "Valid naked domain auto-prepends https",
			input:       "example.com",
			expectedURL: "https://example.com",
			expectedErr: "",
		},
		{
			name:        "Valid domain with existing https prefix preserves scheme",
			input:       "https://salamagi.org",
			expectedURL: "https://salamagi.org",
			expectedErr: "",
		},
		{
			name:        "Valid domain with existing http prefix preserves scheme",
			input:       "http://localhost:8080",
			expectedURL: "http://localhost:8080",
			expectedErr: "",
		},
		{
			name:        "Preserves complex paths and query parameters",
			input:       "example.com/v1/profile?edit=true",
			expectedURL: "https://example.com/v1/profile?edit=true",
			expectedErr: "",
		},
		{
			name:        "Surrounding whitespace is trimmed cleanly",
			input:       "   https://cleaned-up.com   ",
			expectedURL: "https://cleaned-up.com",
			expectedErr: "",
		},
		{
			name:        "Empty string throws requirement error",
			input:       "",
			expectedURL: "",
			expectedErr: "website is required",
		},
		{
			name:        "Whitespace-only string throws requirement error",
			input:       "     ",
			expectedURL: "",
			expectedErr: "website is required",
		},
		{
			name:        "Scheme-only input triggers missing hostname error",
			input:       "https://",
			expectedURL: "",
			expectedErr: "missing hostname",
		},
		{
			name:        "Invalid port signature triggers malformed URL error",
			input:       "example.com:invalidport", // Becomes https://example.com:invalidport which fails parsing
			expectedURL: "",
			expectedErr: "invalid website URL",
		},
	}

	// 2. Iterate through the test collection matrix
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualURL, err := NormalizeWebsiteURL(tt.input)

			// 3. Assert on the error channel state
			if tt.expectedErr != "" {
				// We expect an error to happen
				if err == nil {
					t.Fatalf("NormalizeWebsiteURL(%q) expected error %q, but got nil error", tt.input, tt.expectedErr)
				}
				if err.Error() != tt.expectedErr {
					t.Errorf("NormalizeWebsiteURL(%q) unexpected error message:\n  expected: %q\n  got:      %q", tt.input, tt.expectedErr, err.Error())
				}
			} else {
				// We expect a completely successful execution
				if err != nil {
					t.Fatalf("NormalizeWebsiteURL(%q) unexpected error occurred: %v", tt.input, err)
				}
			}

			// 4. Assert on the string outcome payload
			if actualURL != tt.expectedURL {
				t.Errorf("NormalizeWebsiteURL(%q) unexpected URL string output:\n  expected: %q\n  got:      %q", tt.input, tt.expectedURL, actualURL)
			}
		})
	}
}

func TestFormatNullString(t *testing.T) {
	// 1. Establish the test table covering structural combinations of sql.NullString
	tests := []struct {
		name     string         // Contextual label for the test scenario
		input    sql.NullString // The database null string object wrapper
		expected string         // The expected clean string outcome
	}{
		{
			name: "Valid database record containing text string",
			input: sql.NullString{
				String: "salamagi platform",
				Valid:  true,
			},
			expected: "salamagi platform",
		},
		{
			name: "Valid database record containing an explicit empty string",
			input: sql.NullString{
				String: "",
				Valid:  true,
			},
			expected: "",
		},
		{
			name: "Invalid database record with default zero-value initialization",
			input: sql.NullString{
				String: "",
				Valid:  false,
			},
			expected: "",
		},
		{
			name: "Invalid database record containing an orphaned text value",
			input: sql.NullString{
				String: "this text should be wiped",
				Valid:  false, // 🎯 Tells our app to ignore whatever is sitting in the string block
			},
			expected: "",
		},
	}

	// 2. Iterate through each test case scenario
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := FormatNullString(tt.input)

			// 3. Execute basic assertion check
			if actual != tt.expected {
				t.Errorf(
					"FormatNullString(String: %q, Valid: %v) failed:\n  expected: %q\n  got:      %q",
					tt.input.String, tt.input.Valid, tt.expected, actual,
				)
			}
		})
	}
}

func TestHashResetToken(t *testing.T) {
	// 1. Establish our test matrix
	tests := []struct {
		name        string // Description of the test scenario
		token       string // The input reset token string
		secretKey   string // The mock environment variable value to inject
		expectedHex string // Pre-computed expected result (calculated dynamically below for maintainability)
	}{
		{
			name:      "Standard token with strong secret key",
			token:     "salamagi-reset-token-xyz-123",
			secretKey: "super-secure-dev-secret-key-change-me",
		},
		{
			name:      "Empty token handling",
			token:     "",
			secretKey: "super-secure-dev-secret-key-change-me",
		},
		{
			name:      "Short token with short secret key",
			token:     "a",
			secretKey: "secret",
		},
		{
			name:      "Fallback behavior when secret key environment variable is empty",
			token:     "token-payload",
			secretKey: "", // ⚠️ Cryptographically weak, but tests the function's structural stability
		},
	}

	// 2. Loop through the execution matrix
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 🎯 FIXES EXTERNAL STATE: Safely injects the environment variable for this sub-test only
			t.Setenv("SECRET_KEY", tt.secretKey)

			// Generate the ground-truth hash directly inside the test using standard libraries
			// to cross-verify the utility function's output logic
			h := hmac.New(sha256.New, []byte(tt.secretKey))
			h.Write([]byte(tt.token))
			expectedOutput := hex.EncodeToString(h.Sum(nil))

			// Execute your utility function
			actualOutput := HashToken(tt.token)

			// 3. Assert outputs match perfectly
			if actualOutput != expectedOutput {
				t.Errorf("HashResetToken(%q) failed for SECRET_KEY=%q:\n  expected: %q\n  got:      %q",
					tt.token, tt.secretKey, expectedOutput, actualOutput,
				)
			}
		})
	}
}

func TestHashResetToken_Uniqueness(t *testing.T) {
	// Separate isolated test to guarantee that changing the inputs alters the output
	t.Setenv("SECRET_KEY", "same-secret")

	hashA := HashToken("token-one")
	hashB := HashToken("token-two")

	if hashA == hashB {
		t.Error("Security flaw: Distinct tokens generated identical HMAC hashes.")
	}
}
