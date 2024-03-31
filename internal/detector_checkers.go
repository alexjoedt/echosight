package echosight

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/alexjoedt/echosight/internal/logger"
)

var _ Checker = (*PostgresChecker)(nil)
var _ Validator = (*PostgresChecker)(nil)

type PostgresChecker struct {
	Host        string `json:"host"` // should be come from the Host type, since a dector is always binded to a host
	Port        string `json:"port"`
	Database    string `json:"database"`
	SSL         bool   `json:"ssl"`
	Credentials Credentials
	detector    *Detector `json:"-"`
}

func (p *PostgresChecker) Validate() bool {
	// TODO: implement me
	return true
}

func (p *PostgresChecker) ID() string {
	return p.detector.ID.String()
}

func (p *PostgresChecker) Interval() time.Duration {
	return time.Duration(p.detector.Interval)
}

func (p *PostgresChecker) Detector() *Detector {
	return p.detector
}

func (p *PostgresChecker) Check(ctx context.Context) *Result {
	return &Result{State: StateCritical, Message: "not implemented"}
}

type PingConfig struct {
	detector *Detector `json:"-"`
}

type Credentials struct {
	// TODO: must be stored encrypted in DB
	UsernameCrypt string `json:"username_crypt"`
	PasswordCrypt string `json:"password_crypt"`
	username      string
	password      string
}

func (c *Credentials) Decrypt(crypter Crypter) {
	var err error
	c.username, err = crypter.Decrypt(c.UsernameCrypt)
	if err != nil {
		logger.Fatalf("decrypt username: %v", err)
	}
	c.password, err = crypter.Decrypt(c.UsernameCrypt)
	if err != nil {
		logger.Fatalf("decrypt username: %v", err)
	}
}

func (c *Credentials) BasicAuth() string {
	return base64.RawStdEncoding.EncodeToString([]byte(c.username + ":" + c.password))
}
