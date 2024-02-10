package constants

import "time"

const (
	TofanObjectTemplateFinalizer        = "tofan.io/finalizer"
	RequeueAfter                        = time.Minute * 3
	ObjConditionReady            string = "Ready"
	ObjConditionCreating         string = "Creating"
	ObjConditionFailed           string = "Failed"
)
