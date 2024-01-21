package ctfbot

import (
	"context"
	"time"
)

type CTF struct {
	ID    int
	Name  string
	Start time.Time

	// Discord-related information.
	RoleID  string
	CanJoin bool

	// CTFTime infos.
	CTFTimeURL string

	// Metadata about creation.
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *CTF) Validate() error {
	if c.Name == "" {
		return Errorf(EINVALID, "Name required.")
	}

	if c.RoleID == "" {
		return Errorf(EINVALID, "Player role required.")
	}

	return nil
}

type CTFService interface {
	// Creates a new CTF.
	CreateCTF(ctx context.Context, ctf *CTF) error

	// Retrieves a CTF by name.
	FindCTFByName(ctx context.Context, name string) (*CTF, error)

	// Retrieves a list of ctfs by filter.
	FindCTFs(ctx context.Context, filter CTFFilter) ([]*CTF, int, error)

	// Updates a CTF object.
	UpdateCTF(ctx context.Context, name string, upd CTFUpdate) (*CTF, error)

	// Permanently deletes a CTF.
	DeleteCTF(ctx context.Context, name string) error
}

// CTFFilter represents a filter passed to FindCTFs().
type CTFFilter struct {
	ID      *int
	Name    *string
	RoleID  *string
	CanJoin *bool

	// Limit and offset.
	Limit  int
	Offset int
}

// CTFUpdate represents a filter passed to UpdateCTF().
type CTFUpdate struct {
	Name       *string
	RoleID     *string
	CanJoin    *bool
	CTFTimeURL *string
	Start      *time.Time
}
