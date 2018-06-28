package bloom

const (
	// ErrZeroElements is returned from New() when number of expected elements is zero.
	// It must be at least one.
	ErrZeroElements = Error("number of elements must be positive")
	// ErrProbability is returned from New() when given probability of false-positives
	// is not a positive number. Zero probability doesn't make sense.
	ErrProbability = Error("probability must be positive")
)

// Error defines Bloom filter errors.
type Error string

func (e Error) Error() string {
	return string(e)
}
