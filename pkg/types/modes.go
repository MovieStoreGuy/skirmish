package types

type Mode string

const (
	// DryRun is the string used to defined a dryrun plan
	DryRun = "dryrun"
	// Repairable will ensure that the act of running the orchestrator against the project can be recovered
	Repairable = "repairable"
	// Destruction implies that the state of the project is not important and makes no promises of bring things back
	Destruction = "destruction"
)
