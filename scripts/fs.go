package scripts

import (
	_ "embed"
)

//go:embed seed.sql
var SeedSql string
