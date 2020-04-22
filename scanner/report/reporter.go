package report

import (
	"io"

	"gitlab.com/browserker/browserker"
)

type Reporter struct {
	reports map[string]map[string]*browserker.Report
}

func New() *Reporter {
	return &Reporter{reports: make(map[string]map[string]*browserker.Report, 0)}
}

func (r *Reporter) AddReport(vulnID string, report *browserker.Report) {
	key := report.VulnID + report.Evidence.GetHash()
	if _, exist := r.reports[vulnID]; exist {
		r.reports[vulnID][key] = report
	}
	r.reports[vulnID] = make(map[string]*browserker.Report)
	r.reports[vulnID][key] = report
}

func (r *Reporter) Print(writer io.Writer) {
	return
}
