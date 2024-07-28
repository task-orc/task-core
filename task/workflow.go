package task

import (
	"fmt"
	"sync"
	"time"
)

// WorkflowDef is a struct that represents the workflow definition
// A workflow is created as an instance of the workflow def
type WorkflowDef struct {
	Identity        `json:",inline" bson:",inline"`
	InitialInput    *DataValue `json:"initialInput" bson:"initialInput"`
	nodes           []WorkflowNode
	OnErrorWorkFlow *Workflow `json:"onErrorWorkflow" bson:"onErrorWorkflow"`
}

func NewWorkflowDef(identity Identity, initialInput *DataValue, errorWorkflow *Workflow, nodes ...WorkflowNode) *WorkflowDef {
	if len(identity.ID) == 0 {
		identity = identity.GenerateID("workflow_")
	}
	return &WorkflowDef{
		Identity:        identity,
		InitialInput:    initialInput,
		nodes:           nodes,
		OnErrorWorkFlow: errorWorkflow,
	}
}

// CreateWorkflow creates a new workflow instance from the workflow definition
func (w WorkflowDef) CreateWorkflow() *Workflow {
	result := &Workflow{
		Identity:         w.Identity.GenerateID("workflow_"),
		InitialInput:     w.InitialInput,
		Nodes:            w.nodes,
		OnErrorWorkFlow:  w.OnErrorWorkFlow,
		currentNodeIndex: -1,
		nodeMap:          make(map[string]int, len(w.nodes)),
	}
	for i, node := range w.nodes {
		result.nodeMap[node.GetInfo().ID] = i
		if wfNode, ok := node.(*Workflow); ok {
			// if the node is a workflow, we need to add all the node ids corresponding to the workflow's nodes
			// to be mapped to the index of the workflow node
			// so when status, we can traverse back to the original node using this map, reagrdless of which inner workflow node it is situated at
			//Eg:
			//	workflow1
			//		workflow2
			//			task1
			//			task2
			//		workflow3
			//			task3
			// with node map the data will look like the below
			//	workflow1 { "workflow2": 0, "workflow3": 1, "task1": 0, "task2": 0, "task3": 1 }
			//		workflow2 { "task1": 0, "task2": 1 }
			//			task1
			//			task2
			//		workflow3 { "task3": 0 }
			//			task3

			for _, id := range wfNode.GetAllNodeIds() {
				result.nodeMap[id] = i
			}
		}
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
	isRunning        bool
	sync.Mutex
}

func (w *Workflow) GetAllNodeIds() []string {
	result := []string{}
	for _, node := range w.Nodes {
		// we add all the node ids to the result
		result = append(result, node.GetInfo().ID)
		if wfNode, ok := node.(*Workflow); ok {
			// if it is a workflow, we call the node ids recursively
			result = append(result, wfNode.GetAllNodeIds()...)
		}
	}
	return result
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

// Execute executes the workflow
// It is meant to be called only once
// It will return the status of the workflow, if it has already running
// This also enable the store the state of the workflow and resume it later
func (w *Workflow) Execute(input *DataValue) ExecutionReport {
	if w.isRunning {
		return w.Status()
	}
	w.Lock()
	if w.InitialInput == nil && input != nil {
		w.InitialInput = input
	}
	if w.currentNodeIndex < 0 {
		w.currentNodeIndex = 0
	}
	w.Unlock()
	go func() {
		// we keep listening to the status of the workflow
		rpt := w.Status()
		for !rpt.HasFinished {
			err := w.Run()
			if err != nil {
				break
			}
			time.Sleep(time.Millisecond * 2)
			rpt = w.Status()
		}
	}()
	return w.Status()
}

func (w *Workflow) UpdateStatus(update ExecutionData) {
	if index, ok := w.nodeMap[update.NodeId]; ok {
		w.Nodes[index].UpdateStatus(update)
	}
}

func (w *Workflow) Run() error {
	w.Lock()
	w.isRunning = true
	w.Unlock()
	for w.ShouldMoveForward() {
		node := w.Nodes[w.currentNodeIndex]
		status := node.Status()
		if status.HasFinished && status.Error == nil {
			// Task has finished successfully. Move to next task
			w.Lock()
			w.currentNodeIndex++
			w.Unlock()
			time.Sleep(time.Millisecond * 10)
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
			// If the node is not yet started. Have to start that now
			go node.Execute(w.GetLastNodeOutput())
		} else if status.HasStarted && !status.HasFinished {
			time.Sleep(time.Millisecond * 10)
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
	if w.currentNodeIndex < 0 {
		return nil
	}
	if w.currentNodeIndex == 0 {
		return w.InitialInput
	}
	if w.currentNodeIndex > len(w.Nodes) {
		return nil
	}
	return w.Nodes[w.currentNodeIndex-1].Status().Output
}
