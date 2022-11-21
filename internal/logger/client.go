package logger

// Different logging clients are expected to implement the
// interfaces defined below, can create specific "drivers"
// that route logs to streams other than std
type Client interface {
	Infof(string)
}
