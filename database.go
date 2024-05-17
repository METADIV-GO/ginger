package ginger

func NewMigrate(m ...any) {
	if len(m) == 0 {
		return
	}
	Engine.DBMigrate = append(Engine.DBMigrate, m...)
}

func NewMemMigrate(m ...any) {
	if len(m) == 0 {
		return
	}
	Engine.MemMigrate = append(Engine.MemMigrate, m...)
}
