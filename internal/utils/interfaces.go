package utils

type ApprovalDecision string

const (
	ApprovalYes ApprovalDecision = "yes"
	ApprovalNo  ApprovalDecision = "no"
	ApprovalAll ApprovalDecision = "all"
)

type Reporter interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type Approver interface {
	Approve(action, oldName, newName string) (ApprovalDecision, error)
}

type NopReporter struct{}

func (NopReporter) Infof(string, ...any)  {}
func (NopReporter) Warnf(string, ...any)  {}
func (NopReporter) Errorf(string, ...any) {}
