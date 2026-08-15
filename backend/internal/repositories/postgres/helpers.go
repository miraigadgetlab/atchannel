package postgres

import (
	"errors"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/kosero/atchannel/backend/internal/repositories"
)

var ErrNotFound = repositories.ErrNotFound

func mapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return repositories.ErrNotFound
	}
	return err
}

func pgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func pgTextOrNil(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgUUID(s *string) pgtype.UUID {
	if s == nil {
		return pgtype.UUID{}
	}
	u, err := uuid.Parse(*s)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: u, Valid: true}
}

func pgTime(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func pgBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

func pgInet(ip *netip.Addr) *netip.Addr {
	if ip == nil {
		return nil
	}
	a := *ip
	return &a
}

func ptrStr(s string) *string { return &s }
