package postgres

import (
	"context"
	"crypto/sha256"
	"time"

	es "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/uptrace/bun"
)

var _ es.SessionService = (*SessionModel)(nil)

type SessionModel struct {
	db          *bun.DB
	log         *logger.Logger
	stopCleanup chan bool
}

func (m *SessionModel) Put(ctx context.Context, session *es.Session) error {

	_, err := m.db.NewInsert().Model(session).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m *SessionModel) Get(ctx context.Context, token string) (*es.Session, bool, error) {
	tokenHash := sha256.Sum256([]byte(token))

	var session es.Session

	err := m.db.NewSelect().Model(&session).
		Relation("User").
		Where("hash = ?", tokenHash).Scan(ctx)
	if err != nil {
		return nil, false, err
	}

	return &session, true, nil
}

func (m *SessionModel) Delete(ctx context.Context, token string) error {

	return nil
}

func (m *SessionModel) StartCleanup(interval time.Duration) {
	m.stopCleanup = make(chan bool)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := m.deleteExpired()
			if err != nil {
				m.log.Errorf("%v", err)
			}
		case <-m.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

func (m *SessionModel) StopCleanup() {
	if m.stopCleanup != nil {
		m.stopCleanup <- true
	}
}

func (m *SessionModel) deleteExpired() error {
	ctx := context.Background()
	_, err := m.db.NewDelete().Model(&es.Session{}).Where("expiry < ?", time.Now()).Exec(ctx)
	if err != nil {
		m.log.Errorf("%v", err)
		return err
	}

	return nil
}
