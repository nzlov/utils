package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Map map[string]any

// Scan implements the sql.Scanner interface for Map.
func (n *Map) Scan(value any) error {
	if value == nil {
		*n = nil
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
	return json.Unmarshal(data, n)
}

// Value implements the driver.Valuer interface for Map.
func (n Map) Value() (driver.Value, error) {
	if n == nil {
		return nil, nil
	}
	return json.Marshal(n)
}
