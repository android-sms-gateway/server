package mysql

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

const (
	ErrCodeDuplicateEntry = 1062
)

func IsDuplicateKeyViolation(err error) bool {
	if me, ok := errors.AsType[*mysql.MySQLError](err); ok {
		return me.Number == ErrCodeDuplicateEntry
	}
	return false
}
