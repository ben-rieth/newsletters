package service

import "github.com/ben-rieth/newsletter-api/internal/db"

type UserService struct {
	queries db.Queries
}

func NewUserService(queries db.Queries) *UserService {
	return &UserService{queries}
}
