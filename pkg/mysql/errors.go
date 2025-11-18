package mysql

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

const (
	ErrCodeDuplicateEntry = 1062
)

func IsDuplicateKeyViolation(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == ErrCodeDuplicateEntry
	}
	return false
}
