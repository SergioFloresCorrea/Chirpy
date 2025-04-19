package main

import (
	"time"

	"github.com/SergioFloresCorrea/Chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func MapCreateUserParamsToUser(params database.CreateUserParams) User {
	return User{
		ID:        params.ID,
		CreatedAt: params.CreatedAt,
		UpdatedAt: params.UpdatedAt,
		Email:     params.Email,
	}
}
