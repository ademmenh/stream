package infrastructure

import "go-starter/internal/shared/domain"

type IDGenerator struct{}

func (g *IDGenerator) Generate() domain.Id {
	return domain.NewId()
}
