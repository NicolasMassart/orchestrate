package entities

type GasIncrementLevel string

var (
	GasIncrementVeryLow  GasIncrementLevel = "very-low"
	GasIncrementLow      GasIncrementLevel = "low"
	GasIncrementMedium   GasIncrementLevel = "medium"
	GasIncrementHigh     GasIncrementLevel = "high"
	GasIncrementVeryHigh GasIncrementLevel = "very-high"
)

func (gil *GasIncrementLevel) String() string {
	return string(*gil)
}
