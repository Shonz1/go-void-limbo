package types

type Phase = byte

const (
	PhaseHandshake     Phase = 0
	PhaseStatus        Phase = 1
	PhaseLogin         Phase = 2
	PhaseConfiguration Phase = 3
	PhasePlay          Phase = 4
)
