package camogo

type errors []error

func (e errors) Error() string {
	if len(e) == 1 {
		return e[0].Error()
	}

	// multiple errors
	msg := ""
	needsNewLine := false
	for _, err := range e {
		if needsNewLine {
			msg += "\n"
		}
		msg += err.Error()
		needsNewLine = true
	}
	return msg
}
