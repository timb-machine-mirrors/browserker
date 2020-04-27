package browserker

import "io"

type Evidence struct {
}

func (e *Evidence) Hash() string {
	return ""
}

type Report struct {
	VulnID      string
	CWE         int
	Description string
	Remediation string
	Response    *HTTPResponse
	// TODO: add Navigation type as alternative for flaws that don't have http responses
	Evidence *Evidence
}

type Reporter interface {
	Add(report *Report)
	Print(writer io.Writer)
}
