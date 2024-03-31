package echosight

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	Plaintext *string
	Hash      []byte
}

func (p *Password) Set(plaintextPass string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPass), 12)
	if err != nil {
		return err
	}

	p.Plaintext = &plaintextPass
	p.Hash = hash

	return nil
}

func (p *Password) Matches(plaintextPass string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plaintextPass))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}
