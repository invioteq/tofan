package testcase

const (
	StatusPending    string = "Pending"
	StatusInProgress string = "InProgress"
	StatusCompleted  string = "Completed"
	StatusError      string = "Error"

	StatusPendingReason    string = "AwaitingExecution"
	StatusInProgressReason string = "ExecutionStarted"
	StatusCompletedReason  string = "ExecutionSuccessful"
	StatusErrorReason      string = "ExecutionFailed"

	StatusPendingMsg    string = "The TestCase is pending and has not started execution."
	StatusInProgressMsg string = "The TestCase is currently in progress."
	StatusCompletedMsg  string = "The TestCase has completed successfully."
	StatusErrorMsg      string = "The TestCase encountered an error during execution."
)
