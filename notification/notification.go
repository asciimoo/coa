package notification

import (
	"errors"

	"github.com/asciimoo/coa/event"
)

type Notifier interface {
	Initialize(map[string]string) error
	Notify(*event.Event) error
	Destruct() error
}

type NotifierBackend struct {
	Type string
	Args map[string]string
	notifier Notifier
}

var NotifierTypes = map[string]Notifier{
	"shell": &ShellNotifier{},
	"log": &LogNotifier{},
}

var backends []*NotifierBackend

func Initialize(nbs []*NotifierBackend) error {
	// shut down currently running backends
	if backends != nil {
		for _, nb := range backends {
			// TODO error handling
			nb.notifier.Destruct()
		}
	}

	// initialize new backends
	for _, nb := range nbs {
		if notifier, ok := NotifierTypes[nb.Type]; ok {
			nb.notifier = notifier
			err := nb.notifier.Initialize(nb.Args)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Unknown notifier type: " + nb.Type)
		}
	}
	backends = nbs
	return nil
}

func Send(e *event.Event) error {
	// TODO proper error handling. Now only the last error is shown
    var notifError error
	for _, backend := range backends {
		if backend != nil {
			err := backend.notifier.Notify(e)
			if err != nil {
				notifError = err
			}
		}
	}

	return notifError
}
