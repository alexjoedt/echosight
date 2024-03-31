package postgres

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/filter"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var _ echosight.UserService = (*UserModel)(nil)

type UserModel struct {
	db  *bun.DB
	log *logger.Logger
}

func (m *UserModel) Create(ctx context.Context, user *echosight.User) error {
	user.CreatedAt = time.Now()
	_, err := m.db.NewInsert().
		Model(user).
		Ignore(). // on conflict do nothing
		Exec(ctx)
	if err != nil {
		m.log.Errorc("failed to insert user", err)
		return err
	}

	return nil
}

func (m *UserModel) GetByID(ctx context.Context, id uuid.UUID) (*echosight.User, error) {
	user := new(echosight.User)
	err := m.db.NewSelect().Model(user).
		Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, echosight.ErrNotfoundf("no user found")
		}
		m.log.Errorc("failed to get user by id", err, logger.UUID("user_id", id))
		return nil, err
	}

	m.setPasswordHash(user)
	return user, nil
}

func (m *UserModel) GetByEmail(ctx context.Context, email string) (*echosight.User, error) {
	user := new(echosight.User)
	err := m.db.NewSelect().Model(user).
		Where("email = ?", email).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, echosight.ErrNotfoundf("no user found")
		}
		m.log.Errorc("failed to get user by email", err, logger.Str("email", email))
		return nil, echosight.ErrInternalf("failed to get user by email").WithError(err)
	}

	m.setPasswordHash(user)

	return user, nil
}

func (m *UserModel) Update(ctx context.Context, user *echosight.User) error {
	user.UpdatedAt = time.Now()
	lv := user.LookupVersion
	user.LookupVersion++

	_, err := m.db.NewUpdate().Model(user).
		Where("id = ? AND lookup_version = ?", user.ID, lv).
		Exec(ctx)
	if err != nil {
		m.log.Errorc("failed to update user", err, logger.UUID("user_id", user.ID))
		return echosight.ErrInternalf("failed to update user").WithError(err)
	}

	m.setPasswordHash(user)

	return nil
}

func (m *UserModel) DeleteByID(ctx context.Context, id uuid.UUID) (*echosight.User, error) {
	user := new(echosight.User)
	err := m.db.NewDelete().Model(user).
		Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (m *UserModel) List(ctx context.Context, userFilter *filter.UserFilter) ([]*echosight.User, error) {
	users := make([]*echosight.User, 0)
	query := m.db.NewSelect().Model(&users)

	if userFilter.FirstName != "" {
		query.Where("LOWER(first_name) LIKE ?", "%"+strings.ToLower(userFilter.FirstName)+"%")
	}

	if userFilter.LastName != "" {
		query.Where("LOWER(last_name) LIKE ?", "%"+strings.ToLower(userFilter.LastName)+"%")
	}

	if userFilter.Role != "" {
		query.Where("LOWER(role) LIKE ?", "%"+strings.ToLower(userFilter.Role)+"%")
	}

	if userFilter.Email != "" {
		query.Where("LOWER(email) LIKE ?", "%"+strings.ToLower(userFilter.Email)+"%")
	}

	count, err := query.
		Limit(userFilter.Limit()).
		Offset(userFilter.Offset()).
		Order(userFilter.Order()).
		ScanAndCount(ctx)

	if err != nil {
		return nil, err
	}

	for _, u := range users {
		m.setPasswordHash(u)
	}

	userFilter.Pagination = filter.ComputePagination(count, userFilter.Page, userFilter.PageSize)
	return users, nil
}

func (m *UserModel) setPasswordHash(user *echosight.User) {
	hash, err := base64.StdEncoding.DecodeString(string(user.PasswordHash))
	if err != nil {
		panic(err)
	}

	user.Password = echosight.Password{
		Hash: hash,
	}
}
