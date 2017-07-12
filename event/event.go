package event

type EventType int

const (
    Fail EventType = iota
    Pass
	Notice
	Warning
)

type Event struct {
	Type EventType
	Message string
	ProjectName string
	CheckerName string
}

func EventTypeName(t EventType) string {
	switch t {
	case Fail:
		return "Fail"
	case Pass:
		return "Pass"
	case Notice:
		return "Notice"
	case Warning:
		return "Warning"
	}
	return "[Unknown event type]"
}
