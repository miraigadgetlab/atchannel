package repositories

import "errors"

// ErrNotFound is returned by repositories when a record does not exist.
// It is defined here (not in the postgres package) so mocks and the
// service layer can match it without importing the driver.
var ErrNotFound = errors.New("not found")

// ErrBoardExists is returned when a board with the given slug already exists.
var ErrBoardExists = errors.New("board already exists")
