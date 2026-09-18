package cli

import "testing"

func TestCreateBranchName(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		summary string
		want    string
	}{
		{"plain summary", "JIRA-123", "This is a test summary", "JIRA-123-this-is-a-test-summary"},
		{"special characters stripped", "JIRA-123", "This is a test summ$ry with speci@l chars $%^&*()", "JIRA-123-this-is-a-test-summry-with-specil-chars"},
		{"existing dash kept", "JIRA-123", "Summary with-dash", "JIRA-123-summary-with-dash"},
		{"underscore becomes dash", "JIRA-123", "Summary with_underscore", "JIRA-123-summary-with-underscore"},
		{"non-breaking space", "JIRA-123", "Summary" + string(rune(0xA0)) + "with nbsp", "JIRA-123-summary-with-nbsp"},
		{"braces survive", "JIRA-123", "Summary {1} with braces", "JIRA-123-summary-{1}-with-braces"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createBranchName(tt.key, tt.summary); got != tt.want {
				t.Errorf("createBranchName(%q, %q) = %q, want %q", tt.key, tt.summary, got, tt.want)
			}
		})
	}
}

func TestJiraKeyFromBranchName(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"JIRA-123-this-is-a-test-summary", "JIRA-123"},
		{"JIRA-123-this-is-a-test-summary-with-123-numbers", "JIRA-123"},
		{"JIRA-123", "JIRA-123"},
		{"this-is-a-test-summary", ""},
		{"test-summary-JIRA-123", ""},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			if got := jiraKeyFromBranchName(tt.branch); got != tt.want {
				t.Errorf("jiraKeyFromBranchName(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}

func TestJiraSummaryFromBranchName(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"JIRA-123-this-is-a-test-summary", "This is a test summary"},
		{"JIRA-123-this-is-a-test-summary-with-123-numbers", "This is a test summary with 123 numbers"},
		{"JIRA-123", ""},
		{"this-is-a-test-summary", ""},
		{"test-summary-JIRA-123", ""},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			if got := jiraSummaryFromBranchName(tt.branch); got != tt.want {
				t.Errorf("jiraSummaryFromBranchName(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}

func TestCreateTitleFromBranchName(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		want    string
		wantErr bool
	}{
		{"conventional prefix kept", "feat/add-user-avatar", "feat: Add user avatar", false},
		{"only the first prefix is kept", "feat/fix/mobile/login-flow", "feat: Mobile login flow", false},
		{"underscores become spaces", "fix/crash_on_start", "fix: Crash on start", false},
		{"non-prefix segments are kept", "mobile/feat/login-flow", "Mobile feat login flow", false},
		{"release prefix", "release/1.2.0", "release: 1.2.0", false},
		{"nothing left to title", "/", "", true},
		{"empty branch", "", "", true},
		{"multi-byte first rune", "ärger-im-code", "Ärger im code", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createTitleFromBranchName(tt.branch)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("createTitleFromBranchName(%q) = %q, want error", tt.branch, got)
				}
				if err.Error() != "No title found." {
					t.Errorf("error = %q, want %q", err.Error(), "No title found.")
				}
				return
			}
			if err != nil {
				t.Fatalf("createTitleFromBranchName(%q) returned error: %v", tt.branch, err)
			}
			if got != tt.want {
				t.Errorf("createTitleFromBranchName(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}
