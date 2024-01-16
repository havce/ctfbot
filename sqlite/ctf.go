package sqlite

import (
	"context"
	"strings"

	"github.com/havce/havcebot"
)

type CTFService struct {
	db *DB
}

func NewCTFService(db *DB) *CTFService {
	return &CTFService{
		db: db,
	}
}

func (s *CTFService) FindCTFByName(ctx context.Context, name string) (*havcebot.CTF, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Fetch CTF object.
	return findCTFByName(ctx, tx, name)
}

func (s *CTFService) FindCTFs(ctx context.Context, filter havcebot.CTFFilter) ([]*havcebot.CTF, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	return findCTFs(ctx, tx, filter)
}

func (s *CTFService) CreateCTF(ctx context.Context, ctf *havcebot.CTF) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Create CTF.
	if err := createCTF(ctx, tx, ctf); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *CTFService) UpdateCTF(ctx context.Context, name string, upd havcebot.CTFUpdate) (*havcebot.CTF, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Update the CTF object.
	ctf, err := updateCTF(ctx, tx, name, upd)
	if err != nil {
		return ctf, err
	}
	return ctf, tx.Commit()
}

func (s *CTFService) DeleteCTF(ctx context.Context, name string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := deleteCTF(ctx, tx, name); err != nil {
		return err
	}
	return tx.Commit()
}

func findCTFByName(ctx context.Context, tx *Tx, name string) (*havcebot.CTF, error) {
	ctfs, _, err := findCTFs(ctx, tx, havcebot.CTFFilter{Name: &name})
	if err != nil {
		return nil, err
	} else if len(ctfs) == 0 {
		return nil, havcebot.Errorf(havcebot.ENOTFOUND, "CTF not found.")
	}
	return ctfs[0], nil
}

func findCTFs(ctx context.Context, tx *Tx, filter havcebot.CTFFilter) (_ []*havcebot.CTF, n int, err error) {
	// Build WHERE clause. Each part of the WHERE clause is AND-ed together.
	// Values are appended to an arg list to avoid SQL injection.
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}

	if v := filter.Name; v != nil {
		where, args = append(where, "name = ?"), append(args, *v)
	}

	if v := filter.RoleID; v != nil {
		where, args = append(where, "role_id = ?"), append(args, *v)
	}

	if v := filter.CanJoin; v != nil {
		// Poor man's bool-int conversion.
		canJoin := 0
		if *v {
			canJoin = 1
		}

		where, args = append(where, "can_join"), append(args, canJoin)
	}

	// Execue query with limiting WHERE clause and LIMIT/OFFSET injected.
	rows, err := tx.QueryContext(ctx, `
		SELECT 
		    id,
		    name,
		    start,
		    role_id,
			can_join,
			ctftime_url,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM ctfs
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id ASC
		`+FormatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err != nil {
		return nil, n, FormatError(err)
	}
	defer rows.Close()

	// Iterate over rows and deserialize into CTF objects.
	ctfs := make([]*havcebot.CTF, 0)
	for rows.Next() {
		var ctf havcebot.CTF
		if err := rows.Scan(
			&ctf.ID,
			&ctf.Name,
			(*NullTime)(&ctf.Start),
			&ctf.RoleID,
			&ctf.CanJoin,
			&ctf.CTFTimeURL,
			(*NullTime)(&ctf.CreatedAt),
			(*NullTime)(&ctf.UpdatedAt),
			&n,
		); err != nil {
			return nil, 0, err
		}
		ctfs = append(ctfs, &ctf)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return ctfs, n, nil
}

// createCTF creates a new CTF.
func createCTF(ctx context.Context, tx *Tx, ctf *havcebot.CTF) error {
	// Set timestamps to current time.
	ctf.CreatedAt = tx.now
	ctf.UpdatedAt = ctf.CreatedAt

	// Perform basic field validation.
	if err := ctf.Validate(); err != nil {
		return err
	}

	// Insert row into database.
	result, err := tx.ExecContext(ctx, `
		INSERT INTO ctfs (
			name,
			start,
			role_id,
			can_join,
			ctftime_url,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		ctf.Name,
		(*NullTime)(&ctf.Start),
		ctf.RoleID,
		ctf.CanJoin,
		ctf.CTFTimeURL,
		(*NullTime)(&ctf.CreatedAt),
		(*NullTime)(&ctf.UpdatedAt),
	)
	if err != nil {
		return FormatError(err)
	}

	// Read back new ctf ID into caller argument.
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	ctf.ID = int(id)

	return nil
}

// updateCTF updates a ctf by name. Returns the new state of the ctf after update.
func updateCTF(ctx context.Context, tx *Tx, name string, upd havcebot.CTFUpdate) (*havcebot.CTF, error) {
	// Fetch current object state. Return an error if current user is not owner.
	ctf, err := findCTFByName(ctx, tx, name)
	if err != nil {
		return ctf, err
	}

	// Update fields, if set.
	if v := upd.Name; v != nil {
		ctf.Name = *v
	}

	if v := upd.CanJoin; v != nil {
		ctf.CanJoin = *v
	}

	if v := upd.RoleID; v != nil {
		ctf.RoleID = *v
	}

	if v := upd.CTFTimeURL; v != nil {
		ctf.CTFTimeURL = *v
	}

	if v := upd.Start; v != nil {
		ctf.Start = *v
	}

	ctf.UpdatedAt = tx.now

	// Perform basic field validation.
	if err := ctf.Validate(); err != nil {
		return ctf, err
	}

	// Execute update query.
	if _, err := tx.ExecContext(ctx, `
		UPDATE ctfs
		SET can_join = ?,
			start = ?,
			ctftime_url = ?,
			role_id = ?,
			start = ?,
		    updated_at = ?
		WHERE name = ?
	`,
		ctf.CanJoin,
		(*NullTime)(&ctf.Start),
		ctf.CTFTimeURL,
		ctf.RoleID,
		(*NullTime)(&ctf.Start),
		(*NullTime)(&ctf.UpdatedAt),
		name,
	); err != nil {
		return ctf, FormatError(err)
	}

	return ctf, nil
}

// deleteCTF permanently deletes a CTF by name.
func deleteCTF(ctx context.Context, tx *Tx, name string) error {
	if _, err := findCTFByName(ctx, tx, name); err != nil {
		return err
	}

	// Remove row from database.
	if _, err := tx.ExecContext(ctx, `DELETE FROM ctfs WHERE name = ?`, name); err != nil {
		return FormatError(err)
	}
	return nil
}
