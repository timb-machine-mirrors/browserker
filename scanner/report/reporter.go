package report

import (
	"io"

	"gitlab.com/browserker/browserk"
)

type Reporter struct {
	reports map[string]map[string]*browserk.Report
}

func New() *Reporter {
	return &Reporter{reports: make(map[string]map[string]*browserk.Report, 0)}
}

func (r *Reporter) Add(report *browserk.Report) {
	key := report.VulnID + report.Evidence.Hash()
	if _, exist := r.reports[report.VulnID]; exist {
		r.reports[report.VulnID][key] = report
	}
	r.reports[report.VulnID] = make(map[string]*browserk.Report)
	r.reports[report.VulnID][key] = report
}

func (r *Reporter) Print(writer io.Writer) {
	return
}
