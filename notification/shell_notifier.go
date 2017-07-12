package notification

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"text/template"

	"github.com/asciimoo/coa/event"
)


// Shell notifier executes commands on fail and pass events.
// Event message is acessible from the commands with the %q format string.
// Arguments:
//    "fail_command": command to run on fail event
//    "pass_command": command to run on pass event
//
// TODO implement proper shell escaping instead of relying on golang's format strings
type ShellNotifier struct {
	failCmdTpl *template.Template
	passCmdTpl *template.Template
}

func (s *ShellNotifier) Initialize(args map[string]string) error {
	if failFormat, ok := args["fail_command"]; ok {
		s.failCmdTpl, _ = template.New("fail template").Parse(failFormat)
	}
	if passFormat, ok := args["pass_command"]; ok {
		s.passCmdTpl, _ = template.New("pass template").Parse(passFormat)
	}
	if s.failCmdTpl == nil && s.passCmdTpl == nil {
		return errors.New("No (pass|fail)_command specified for shell notifier")
	}
	return nil
}

type shellArgs struct {
	Title string
	Message string
}

func (s *ShellNotifier) Notify(e *event.Event) error {
	// TODO use proper shell escaping instead of %q format string
	title := fmt.Sprintf("%q", fmt.Sprintf("[%v/%v]", e.ProjectName, e.CheckerName))
	msg := fmt.Sprintf("%q", e.Message)
	buf := bytes.NewBufferString("")
	var err error
	switch e.Type {
	case event.Fail:
		if s.failCmdTpl == nil {
			return nil
		}
		err = s.failCmdTpl.Execute(buf, &shellArgs{Title: title, Message: msg})
	case event.Pass:
		if s.passCmdTpl == nil {
			return nil
		}
		err = s.passCmdTpl.Execute(buf, &shellArgs{Title: title, Message: msg})
	default:
		return errors.New("Unsupported event type for shell notifier")
	}

	if err != nil {
		return err
	}

	cmdStr := buf.String()
	if cmdStr == "" {
		return nil
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	return cmd.Run()
}


func (_ *ShellNotifier) Destruct() error {
	return nil
}
