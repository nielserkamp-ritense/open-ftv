package config

// Migration contains the configuration details for database migrations.
//
// Note that the URL for the target database must be retrieved from the Persistence configuration.
// Currently only PostgreSQL is supported.
type Migration struct {
	Source    string `yaml:"migrate.source,omitempty" env:"MIGRATE_SOURCE"   flag:"migrate-source"   desc:"Source location of migration scripts"`
	Steps     int    `yaml:"migrate.steps,omitempty"  env:"MIGRATE_STEPS"    flag:"migrate-steps"    desc:"Migration steps to perform (positive is up, negative is down)"`
	Auto      bool   `yaml:"migrate.auto,omitempty"   env:"MIGRATE_AUTO"     flag:"migrate-auto"     desc:"Always migrate up to the latest level"`
	ExitAfter bool   `yaml:"migrate.exit,omitempty"   env:"MIGRATE_AND_EXIT" flag:"migrate-and-exit" desc:"Exit process after migrations finished"`
}
