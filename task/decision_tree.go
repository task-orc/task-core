package task

import "sync"

type DecisionTree struct {
	Identity
	Task
	PossibleSteps []DecisionTreeStep
	sync.Mutex
}

type DecisionTreeStep struct {
	Condition string
	Step      Workflow
}

func (d *DecisionTree) Execute(input *DataValue) ExecutionReport {
	if d.HasStarted {
		return d.Status()
	}
	d.Lock()
	d.HasStarted = true
	//TODO: Validate input of task with task definition
	// t.Input = input
	d.Unlock()
	exeData := d.execFn(d.ID, input)
	d.Lock()
	d.Output = exeData.Output
	d.Error = exeData.Error
	d.HasFinished = true
	d.Unlock()
	return d.Status()
}
