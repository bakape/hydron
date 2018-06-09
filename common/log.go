package common

// Logger provides callbacks for displaying an operation's progress
type Logger interface {
	SetTotal(int)
	Done(Record)
	Err(error)
}
