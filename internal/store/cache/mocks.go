package cache

import (
	"context"
	"social/internal/store"
	//ex 63 spies package
)

//ex 62 created mockstore for cache and MockUserstore with its interface methods
//"github.com/stretchr/testify/mock" //ex 63 testify spies package read it personally, good package for tests in golang

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct {
	//mock.Mock
}

func (m MockUserStore) Get(ctx context.Context, id int64) (*store.User, error) {
	return nil, nil
}

func (m MockUserStore) Set(ctx context.Context, user *store.User) error {
	return nil
}
