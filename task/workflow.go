package task

import (
	"fmt"
	"sync"
)

type WorkflowDef struct {
	Identity        `json:",inline"`
	InitialInput    *DataValue     `json:"initialInput"`
	Nodes           []WorkflowNode `json:"nodes"`
	OnErrorWorkFlow *Workflow      `json:"onErrorWorkflow"`
	Error           string         `json:"error"`
}

func NewWorkflowDef(identity Identity, initialInput *DataValue, errorWorkflow *Workflow, nodes ...WorkflowNode) *WorkflowDef {
	if len(identity.ID) == 0 {
		identity = identity.GenerateID("workflow_")
	}
	return &WorkflowDef{
		Identity:        identity,
		InitialInput:    initialInput,
		Nodes:           nodes,
		OnErrorWorkFlow: errorWorkflow,
	}
}

func (w WorkflowDef) CreateWorkflow() *Workflow {
	result := &Workflow{
		Identity:         w.Identity.GenerateID("workflow_"),
		InitialInput:     w.InitialInput,
		Nodes:            w.Nodes,
		OnErrorWorkFlow:  w.OnErrorWorkFlow,
		currentNodeIndex: -1,
		nodeMap:          make(map[string]int, len(w.Nodes)),
	}
	for i, node := range w.Nodes {
		result.nodeMap[node.GetInfo().ID] = i
	}
	return result
}

type Workflow struct {
	Identity
	InitialInput     *DataValue
	Nodes            []WorkflowNode
	OnErrorWorkFlow  *Workflow
	currentNodeIndex int
	Error            error
	nodeMap          map[string]int
	sync.Mutex
}

func (w *Workflow) Status() ExecutionReport {
	return ExecutionReport{
		HasStarted:  w.currentNodeIndex >= 0,
		HasFinished: w.Error != nil || w.currentNodeIndex >= len(w.Nodes),
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
	if w.InitialInput == nil && input != nil {
		w.InitialInput = input
	}
	w.currentNodeIndex = 0
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
		node := w.Nodes[w.currentNodeIndex]
		status := node.Status()
		if status.HasFinished && status.Error == nil {
			// Task has finished successfully. Move to next task
			w.Lock()
			w.currentNodeIndex++
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
	if w.currentNodeIndex >= len(w.Nodes) {
		return false
	}
	if w.currentNodeIndex < 0 {
		return false
	}
	return true
}

func (w *Workflow) GetLastNodeOutput() *DataValue {
	if w.currentNodeIndex == 0 {
		return w.InitialInput
	}
	if w.currentNodeIndex > len(w.Nodes) {
		return nil
	}
	return w.Nodes[w.currentNodeIndex-1].Status().Output
}
