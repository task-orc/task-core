package task

import (
	"fmt"
	"sync"
)

type Workflow struct {
	Identity
	InitialInput     *DataValue
	Nodes            []WorkflowNode
	CurrentNodeIndex int
	Error            error
	sync.Mutex
}

func NewWorkflow(identity Identity, nodes ...WorkflowNode) *Workflow {
	return &Workflow{
		Identity:         identity,
		Nodes:            nodes,
		CurrentNodeIndex: -1,
	}
}

func (w *Workflow) Status() ExecutionData {
	return ExecutionData{
		HasStarted:  w.CurrentNodeIndex >= 0,
		HasFinished: w.CurrentNodeIndex >= len(w.Nodes),
		Input:       w.InitialInput,
		Output:      w.GetLastNodeOutput(),
		Error:       w.Error,
	}
}

func (w *Workflow) Execute(input *DataValue) ExecutionData {
	w.Lock()
	w.InitialInput = input
	w.CurrentNodeIndex = 0
	w.Unlock()
	go w.Run()
	return w.Status()
}

func (w *Workflow) Run() error {
	for w.ShouldMoveForward() {
		node := w.Nodes[w.CurrentNodeIndex]
		status := node.Status()
		if status.HasFinished && status.Error == nil {
			// Task has finished successfully. Move to next task
			w.Lock()
			w.CurrentNodeIndex++
			w.Unlock()
			continue
		} else if status.Error != nil {
			// Task has failed. Return error
			w.Lock()
			w.Error = fmt.Errorf("error in node execution %s: %w", node.GetInfo().ID, status.Error)
			w.Unlock()
			return w.Error
		} else if !status.HasStarted {
			// Task is not started. Have to start that now
			go node.Execute(w.GetLastNodeOutput())
		} else if status.HasStarted && !status.HasFinished {
			// Task is in progress
			return nil
		}
	}
	return nil
}

func (w *Workflow) ShouldMoveForward() bool {
	if w.CurrentNodeIndex >= len(w.Nodes) {
		return false
	}
	if w.CurrentNodeIndex < 0 {
		return false
	}
	return true
}

func (w *Workflow) GetLastNodeOutput() *DataValue {
	if w.CurrentNodeIndex == 0 {
		return w.InitialInput
	}
	if w.CurrentNodeIndex > len(w.Nodes) {
		return nil
	}
	return w.Nodes[w.CurrentNodeIndex-1].Status().Output
}
