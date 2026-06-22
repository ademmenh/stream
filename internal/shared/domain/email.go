package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if !emailPattern.MatchString(normalized) {
		return Email{}, fmt.Errorf("invalid email format: %q", raw)
	}
	return Email{value: normalized}, nil
}

func (e Email) Value() string  { return e.value }
func (e Email) String() string { return e.value }

func (e Email) Equal(other Email) bool {
	return e.value == other.value
}

func (e Email) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, e.value)), nil
}

func (e *Email) UnmarshalJSON(b []byte) error {
	s := string(b)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		e.value = s[1 : len(s)-1]
		return nil
	}
	return fmt.Errorf("invalid Email JSON: %s", s)
}
