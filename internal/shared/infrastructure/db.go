package infrastructure

import (
	"fmt"

	_ "github.com/lib/pq"

	"go-starter/internal/shared/infrastructure/ent/generated"
)

func NewDB(dsn string) (*generated.Client, error) {
	client, err := generated.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open ent client: %w", err)
	}
	return client, nil
}
