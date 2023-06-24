package task

type WorkflowNode interface {
	IdentityInfo
	Execution
}

type Execution interface {
	Execute(input *DataValue) ExecutionReport
	Status() ExecutionReport
	UpdateStatus(update ExecutionData)
}

type ExecutionData struct {
	Output *DataValue
	Error  error
	NodeId string
}

type ExecutionReport struct {
	HasStarted  bool
	HasFinished bool
	Input       *DataValue
	ExecutionData
}

type ExecutionFn func(input *DataValue) ExecutionData
