package download

type Status int

const (
	Pending Status = iota + 1
	InProgress
	Paused
	Completed
	Failed
)

type Metadata struct {
	ContentSize      int64
	IsRangeSupported bool
}
