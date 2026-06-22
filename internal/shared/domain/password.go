package domain

import (
	"fmt"
	"unicode/utf8"
)

type Password struct {
	value string
}

func NewPassword(raw string) (Password, error) {
	if utf8.RuneCountInString(raw) < 6 {
		return Password{}, fmt.Errorf("password must be at least 6 characters")
	}
	return Password{value: raw}, nil
}

func (p Password) Value() string  { return p.value }
func (p Password) String() string { return p.value }
