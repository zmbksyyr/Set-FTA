package userchoice

import "testing"

func TestHash(t *testing.T) {
	tests := []struct {
		extension string
		sid       string
		progID    string
		timestamp string
		want      string
	}{
		{".txt", "s-1-5-21-100", "portableassoc.editor", "01dc000000000000", "UBxHti9ONDw="},
		{".mp4", "s-1-5-21-123456", "portableassoc.media", "01db123456789abc", "f1V1ojzGn/k="},
		{".zip", "s-1-5-18", "portableassoc.archive", "01dcffffffffffff", "oILY/TFOTcA="},
	}
	for _, test := range tests {
		if got := Hash(test.extension, test.sid, test.progID, test.timestamp); got != test.want {
			t.Errorf("Hash(%q, %q) = %q, want %q", test.extension, test.progID, got, test.want)
		}
	}
}
