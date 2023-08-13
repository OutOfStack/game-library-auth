package scripts

import (
	_ "embed" // required for embedding
)

// SeedSQL - embedded seed sql content
//
//go:embed seed.sql
var SeedSQL string
