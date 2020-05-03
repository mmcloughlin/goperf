package units

// ImprovementDirection indicates which direction of change is considered better.
type ImprovementDirection int

// Supported improvement direction values.
const (
	ImprovementDirectionUnknown ImprovementDirection = 0
	ImprovementDirectionLarger  ImprovementDirection = 1
	ImprovementDirectionSmaller ImprovementDirection = -1
)

//go:generate enumer -type ImprovementDirection -output direction_enum.go -trimprefix ImprovementDirection -transform snake

// ImprovementDirectionForUnit returns the improvement direction for the supplied unit, if known.
func ImprovementDirectionForUnit(unit string) ImprovementDirection {
	switch unit {
	case Runtime, BytesAllocated, Allocs:
		return ImprovementDirectionSmaller
	case DataRate:
		return ImprovementDirectionLarger
	default:
		return ImprovementDirectionUnknown
	}
}
