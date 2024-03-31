package cache

import (
	"context"
	"encoding/json"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/filter"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/google/uuid"
)

var _ echosight.UserService = (*userCache)(nil)

type userCache struct {
	cache echosight.Cache
	next  echosight.UserService
	log   *logger.Logger
	ttl   time.Duration
}

func NewUserCache(cache echosight.Cache, ttl time.Duration, us echosight.UserService) echosight.UserService {
	return &userCache{
		next:  us,
		cache: cache,
		log:   logger.New("User-Cache"),
		ttl:   ttl,
	}
}

func (u *userCache) Create(ctx context.Context, user *echosight.User) error {

	// first we have to  create the user in the databse because the database assigns the ID
	err := u.next.Create(ctx, user)
	if err != nil {
		return err
	}

	data, err := json.Marshal(user)
	if err != nil {
		u.log.Warnf("failed to marshal user for cache: %v", err)
		return nil
	}

	err = u.cache.Put(context.Background(), "users", user.ID.String(), data, u.ttl)
	if err != nil {
		return err
	}
	return nil
}

func (u *userCache) GetByID(ctx context.Context, id uuid.UUID) (*echosight.User, error) {
	data, err := u.cache.Get(ctx, "users", id.String())
	if err != nil {
		// TODO: handle ErrNotExists or so
		u.log.Debugf("failed to get user from cache: %v", err)
		user, err := u.next.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		u.Create(ctx, user)
		return user, err
	}

	var user echosight.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		u.log.Warnf("failed to unmarshal user from cache: %v", err)
		return u.next.GetByID(ctx, id)
	}

	return &user, nil
}

func (u *userCache) GetByEmail(ctx context.Context, email string) (*echosight.User, error) {
	// in most cases get by email is used to login and there should be no user in the cache when login.
	// So we want always the origin data from the database
	return u.next.GetByEmail(ctx, email)
}

func (u *userCache) Update(ctx context.Context, user *echosight.User) error {
	err := u.next.Update(ctx, user)
	if err != nil {
		return err
	}

	data, err := json.Marshal(user)
	if err != nil {
		u.log.Warnf("failed to marshal user for cache: %v", err)
		return nil
	}

	err = u.cache.Update(ctx, "users", user.ID.String(), data, u.ttl)
	if err != nil {
		u.log.Warnf("failed to update cache %v", err)
	}

	return nil
}

func (u *userCache) DeleteByID(ctx context.Context, id uuid.UUID) (*echosight.User, error) {
	u.cache.Delete(ctx, "users", id.String())
	return u.next.DeleteByID(ctx, id)
}

func (u *userCache) List(ctx context.Context, userFilter *filter.UserFilter) ([]*echosight.User, error) {
	// no caching for list users
	return u.next.List(ctx, userFilter)
}
