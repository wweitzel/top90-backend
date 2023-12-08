package db

import (
	"database/sql/driver"
	"errors"
)

type NullString string

func (s *NullString) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	strVal, ok := value.(string)
	if !ok {
		return errors.New("column is not a string")
	}
	*s = NullString(strVal)
	return nil
}
func (s NullString) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return string(s), nil
}

type NullInt int

func (i *NullInt) Scan(value interface{}) error {
	if value == nil {
		*i = 0
		return nil
	}
	intVal, ok := value.(int64)
	if !ok {
		return errors.New("column is not an int64")
	}
	*i = NullInt(intVal)
	return nil
}
func (i NullInt) Value() (driver.Value, error) {
	if i == 0 {
		return nil, nil
	}
	return int64(i), nil
}
