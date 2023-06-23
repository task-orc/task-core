package task

type WorkflowNode interface {
	IdentityInfo
	Execution
}

type Execution interface {
	Execute(input *DataValue) ExecutionData
	Status() ExecutionData
}

type ExecutionData struct {
	HasStarted  bool
	HasFinished bool
	Input       *DataValue
	Output      *DataValue
	Error       error
	NodeId      string
}
