package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type Id struct {
	value string
}

func NewId() Id {
	return Id{value: uuid.New().String()}
}

func IdFromStr(v string) Id {
	return Id{value: v}
}

func (id Id) Value() string  { return id.value }
func (id Id) String() string { return id.value }

func (id Id) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, id.value)), nil
}

func (id *Id) UnmarshalJSON(b []byte) error {
	s := string(b)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		id.value = s[1 : len(s)-1]
		return nil
	}
	return fmt.Errorf("invalid Id: %s", s)
}
