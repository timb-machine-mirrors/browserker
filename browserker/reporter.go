package browserker

import "io"

type Evidence struct {
}

func (e *Evidence) GetHash() string {
	return ""
}

type Report struct {
	VulnID      string
	CWE         int
	Description string
	Remediation string
	Response    *HttpResponse
	// TODO: add Navigation type as alternative for flaws that don't have http responses
	Evidence *Evidence
}

type Reporter interface {
	AddReport(vulnID string, report *Report)
	Print(writer io.Writer)
}
