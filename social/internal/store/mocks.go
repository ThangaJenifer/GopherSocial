package store

import (
	"context"
	"database/sql"
	"time"
)

//ex 62, we have create a mockstore with implementing all the methods of Storage interfaces which is returning nothing

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct{}

func (m *MockUserStore) Create(context.Context, *sql.Tx, *User) error {
	return nil
}

func (m *MockUserStore) GetByID(context.Context, int64) (*User, error) {
	return &User{}, nil
}
func (m *MockUserStore) GetByEmail(context.Context, string) (*User, error) {
	return &User{}, nil
}

func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return nil
}

func (m *MockUserStore) Activate(context.Context, string) error {
	return nil
}

func (m *MockUserStore) Delete(context.Context, int64) error {
	return nil
}
