package usecase

import "testing"

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{
			name:      "valid password",
			password:  "Secret123",
			wantError: false,
		},
		{
			name:      "too short",
			password:  "Sec123",
			wantError: true,
		},
		{
			name:      "no digit",
			password:  "SecretPassword",
			wantError: true,
		},
		{
			name:      "no letter",
			password:  "12345678",
			wantError: true,
		},
		{
			name:      "cyrillic letters and digits",
			password:  "Пароль123",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errText := validatePassword(tt.password)

			if tt.wantError && errText == "" {
				t.Fatal("expected validation error, got empty string")
			}

			if !tt.wantError && errText != "" {
				t.Fatalf("expected no validation error, got %q", errText)
			}
		})
	}
}
