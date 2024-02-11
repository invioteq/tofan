package constants

import "time"

const (
	TofanFinalizer = "tofan.io/finalizer"
	RequeueAfter   = time.Minute * 3

	ObjConditionReady    string = "Ready"
	ObjConditionCreating string = "Creating"
	ObjConditionFailed   string = "Failed"

	TofanTestCaseNameLabel string = "tofan.io/testcase-name"
)
