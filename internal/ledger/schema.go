package ledger

import "embed"

const (
	SQLiteSchemaName   = "schema/sqlite.sql"
	PostgresSchemaName = "schema/postgres.sql"
)

//go:embed schema/*.sql
var schemaFS embed.FS

// LoadSchema reads a schema file embedded in the binary.
func LoadSchema(name string) (string, error) {
	data, err := schemaFS.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
