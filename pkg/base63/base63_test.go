package base63

import (
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name    string
		input   uint64
		want    string
		wantErr bool
	}{
		{
			name:    "zero_value",
			input:   0,
			want:    "aaaaaaaaaa",
			wantErr: false,
		},
		{
			name:    "small_value",
			input:   1,
			want:    "baaaaaaaaa",
			wantErr: false,
		},
		{
			name:    "alphabet_boundary",
			input:   62,
			want:    "_aaaaaaaaa",
			wantErr: false,
		},
		{
			name:    "multi_digit_transition",
			input:   63,
			want:    "abaaaaaaaa",
			wantErr: false,
		},
		{
			name:    "max_valid_id",
			input:   7645370045518097,
			want:    "1zDEk77YEa",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Encode() got = %s, want %s", got, tt.want)
			}
			
			if !tt.wantErr && len(got) != 10 {
				t.Errorf("Encode() length = %d, want 10", len(got))
			}
		})
	}
}

func TestEncode_Overflow(t *testing.T) {
	// 63^10 + 1 = 98,461,329,465,662,061 + 1
	var tooBigID uint64 = 98461329465662062 
	
	_, err := Encode(tooBigID)
	if err == nil {
		t.Error("expected error for ID exceeding 10 characters, got nil")
	}
}