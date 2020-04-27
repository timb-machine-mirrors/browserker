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

func (r *Reporter) Add(report *browserker.Report) {
	key := report.VulnID + report.Evidence.Hash()
	if _, exist := r.reports[report.VulnID]; exist {
		r.reports[report.VulnID][key] = report
	}
	r.reports[report.VulnID] = make(map[string]*browserker.Report)
	r.reports[report.VulnID][key] = report
}

func (r *Reporter) Print(writer io.Writer) {
	return
}
