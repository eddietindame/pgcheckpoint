package checkpoint

import "fmt"

// NamingMode controls how checkpoint files are named.
type NamingMode struct {
	value string
}

var (
	NamingModeSequential = NamingMode{"sequential"}
	NamingModeTimestamp  = NamingMode{"timestamp"}
	NamingModeCompact    = NamingMode{"compact"}
	NamingModeUnix       = NamingMode{"unix"}
)

// validNamingModes maps string values to their NamingMode.
var validNamingModes = map[string]NamingMode{
	"sequential": NamingModeSequential,
	"timestamp":  NamingModeTimestamp,
	"compact":    NamingModeCompact,
	"unix":       NamingModeUnix,
}

// ParseNamingMode parses a string into a NamingMode, returning an error for invalid values.
func ParseNamingMode(s string) (NamingMode, error) {
	if m, ok := validNamingModes[s]; ok {
		return m, nil
	}
	return NamingMode{}, fmt.Errorf("invalid naming mode %q: must be one of sequential, timestamp, compact, unix", s)
}

// String returns the string representation of the naming mode.
func (m NamingMode) String() string {
	return m.value
}

// IsTimestampBased returns true for modes that use time-based filenames.
func (m NamingMode) IsTimestampBased() bool {
	return m == NamingModeTimestamp || m == NamingModeCompact || m == NamingModeUnix
}
