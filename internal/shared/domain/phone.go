package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var phonePattern = regexp.MustCompile(`^\+?[\d\s\-\(\)]{7,20}$`)

type Phone struct {
	value string
}

func NewPhone(raw string) (Phone, error) {
	normalized := strings.TrimSpace(raw)
	if !phonePattern.MatchString(normalized) {
		return Phone{}, fmt.Errorf("invalid phone number format: %q", raw)
	}
	return Phone{value: normalized}, nil
}

func (p Phone) Value() string  { return p.value }
func (p Phone) String() string { return p.value }

func (p Phone) Equal(other Phone) bool {
	return p.value == other.value
}

func (p Phone) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, p.value)), nil
}

func (p *Phone) UnmarshalJSON(b []byte) error {
	s := string(b)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		p.value = s[1 : len(s)-1]
		return nil
	}
	return fmt.Errorf("invalid Phone JSON: %s", s)
}
