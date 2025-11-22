package users

import "time"

type User struct {
	ID string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func newUser(model *userModel) *User {
	return &User{
		ID: model.ID,

		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
