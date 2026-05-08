package auth

import "testing"

func TestIsGitHubEmailAllowed(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		allowedEmails string
		want          bool
	}{
		{
			name:          "empty allowlist permits any email",
			email:         "user@example.com",
			allowedEmails: "",
			want:          true,
		},
		{
			name:          "email in comma separated allowlist is permitted",
			email:         "second@example.com",
			allowedEmails: "first@example.com,second@example.com",
			want:          true,
		},
		{
			name:          "email match ignores spaces and case",
			email:         "Second@Example.com",
			allowedEmails: " first@example.com, second@example.com ",
			want:          true,
		},
		{
			name:          "email outside allowlist is rejected",
			email:         "third@example.com",
			allowedEmails: "first@example.com,second@example.com",
			want:          false,
		},
		{
			name:          "empty github email is rejected when allowlist is set",
			email:         "",
			allowedEmails: "first@example.com",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGitHubEmailAllowed(tt.email, tt.allowedEmails)
			if got != tt.want {
				t.Fatalf("isGitHubEmailAllowed(%q, %q) = %t, want %t", tt.email, tt.allowedEmails, got, tt.want)
			}
		})
	}
}
