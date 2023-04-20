package datatrans

// cf. https://api-reference.datatrans.ch/#tag/v1transactions/operation/status
type Status string

const (
	StatusInitialized   = "initialized"
	StatusAuthenticated = "authenticated"
	StatusAuthorized    = "authorized"
	StatusSettled       = "settled"
	StatusTransmitted   = "transmitted"
	StatusFailed        = "failed"
	StatusCanceled      = "canceled"
)

var (
	// AllStates represents the list of all valid types
	AllStates = []Status{
		StatusInitialized,
		StatusAuthenticated,
		StatusAuthorized,
		StatusSettled,
		StatusTransmitted,
		StatusFailed,
		StatusCanceled,
	}
)

// String returns the string representation
func (s Status) String() string {
	return string(s)
}

// Valid check if the given value is included
func (s Status) Valid() bool {
	for _, v := range AllStates {
		if v == s {
			return true
		}
	}
	return false
}

// Is returns true if status type equals x
func (s Status) Is(x Status) bool {
	return x != "" && x == s
}
