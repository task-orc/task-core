package task

import (
	"fmt"
	"sync"
)

type Workflow struct {
	Identity
	InitialInput     *DataValue
	Nodes            []WorkflowNode
	OnErrorWorkFlow  *Workflow
	CurrentNodeIndex int
	Error            error
	nodeMap          map[string]int
	sync.Mutex
}

func NewWorkflow(identity Identity, errorWorkflow *Workflow, nodes ...WorkflowNode) *Workflow {
	workflowMap := make(map[string]int, len(nodes))
	for i, node := range nodes {
		workflowMap[node.GetInfo().ID] = i
	}
	if len(identity.ID) == 0 {
		identity = identity.GenerateID("workflow_")
	}
	return &Workflow{
		Identity:         identity,
		Nodes:            nodes,
		OnErrorWorkFlow:  errorWorkflow,
		CurrentNodeIndex: -1,
		nodeMap:          workflowMap,
	}
}

func (w *Workflow) Status() ExecutionReport {
	return ExecutionReport{
		HasStarted:  w.CurrentNodeIndex >= 0,
		HasFinished: w.Error != nil || w.CurrentNodeIndex >= len(w.Nodes),
		Input:       w.InitialInput,
		ExecutionData: ExecutionData{
			Output: w.GetLastNodeOutput(),
			Error:  w.Error,
			NodeId: w.ID,
		},
	}
}

func (w *Workflow) Execute(input *DataValue) ExecutionReport {
	w.Lock()
	w.InitialInput = input
	w.CurrentNodeIndex = 0
	w.Unlock()
	go w.Run()
	return w.Status()
}

func (w *Workflow) UpdateStatus(update ExecutionData) {
	if index, ok := w.nodeMap[update.NodeId]; ok {
		w.Nodes[index].UpdateStatus(update)
	}
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
			w.Error = fmt.Errorf("error in node execution %s - %s: %w", node.GetInfo().ID, node.GetInfo().Name, status.Error)
			if w.OnErrorWorkFlow != nil {
				go w.OnErrorWorkFlow.Execute(w.GetLastNodeOutput())
			}
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
