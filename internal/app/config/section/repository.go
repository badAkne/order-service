package section

import (
	"fmt"
	"time"
)

type Repository struct {
	Postgres RepositoryPostgres
}

type RepositoryPostgres struct {
	Host         string        `required:"true"`
	Port         string        `required:"true"`
	Username     string        `required:"true"`
	Password     string        `required:"true"`
	Name         string        `required:"true"`
	ReadTimeout  time.Duration `default:"30s"`
	WriteTimeout time.Duration `default:"30s"`
}

func (p *RepositoryPostgres) DSN() string {

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		p.Host,
		p.Username,
		p.Password,
		p.Name,
		p.Port)

}
