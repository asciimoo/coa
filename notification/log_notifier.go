package notification

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/asciimoo/coa/event"
)


// Log notifier dumps events to the specified file destination.
// Arguments:
//    "destination": log file path or STDOUT or STDERR (optional, default is STDERR)
type LogNotifier struct {
	destination io.WriteCloser
}

func (l *LogNotifier) Initialize(args map[string]string) error {
	if destination, ok := args["destination"]; ok {
		switch destination {
		case "STDOUT":
			l.destination = os.Stdout
		case "STDERR":
			l.destination = os.Stderr
		default:
			logFile, err := os.OpenFile(destination, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600);
			if err != nil {
				return err
			}
			l.destination = logFile
		}
	} else {
		l.destination = os.Stderr
	}
	return nil
}

func (l *LogNotifier) Notify(e *event.Event) error {
	_, err := fmt.Fprintf(l.destination, "[%v/%v] %v: %v\n",
		e.ProjectName,
		e.CheckerName,
		event.EventTypeName(e.Type),
		strings.Replace(e.Message, "\n", string('\u23CE'), -1),
	)
	return err
}

func (l *LogNotifier) Destruct() error {
	if l.destination == os.Stderr || l.destination == os.Stdout {
		return nil
	}
	return l.destination.Close()
}
