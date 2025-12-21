package ledger

import "testing"

func TestLoadSchema(t *testing.T) {
	sqlite, err := LoadSchema(SQLiteSchemaName)
	if err != nil {
		t.Fatalf("load sqlite schema: %v", err)
	}
	if sqlite == "" {
		t.Fatalf("sqlite schema is empty")
	}

	pg, err := LoadSchema(PostgresSchemaName)
	if err != nil {
		t.Fatalf("load postgres schema: %v", err)
	}
	if pg == "" {
		t.Fatalf("postgres schema is empty")
	}
}
