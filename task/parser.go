package task

import (
	"encoding/json"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

type ParseOutput interface {
	WorkflowNodesDef
}

// Parser helps parsing the data from different sources to the expected ParseOutputs
type Parser[o ParseOutput] interface {
	Parse() (*o, error)
}

type workflowNodeType string

const (
	// Supported concrete parse output types
	// The orginal parse output is an interface, but we need to know the concrete type to unmarshal the json
	workflowNodeTypeTask     workflowNodeType = "task"
	workflowNodeTypeWorkflow workflowNodeType = "workflow"
)

// Concrete implementation for parsing task defnition
// task is created from the instance of the task def
type taskDefParserWithoutExeFn struct {
	TaskDef `json:",inline" bson:",inline"`
	Type    string `json:"type" bson:"type"`
	ExecFn  string `json:"execFn" bson:"execFn"`
}

// UnmarshalJSON is a custom unmarshaler for taskDefParserWithoutExeFn
func (t *taskDefParserWithoutExeFn) UnmarshalJSON(data []byte) error {
	tsk := struct {
		TaskDef `json:",inline"`
		Type    string `json:"type"`
		ExecFn  string `json:"execFn"`
	}{}
	err := json.Unmarshal(data, &tsk)
	if err != nil {
		return err
	}
	if tsk.Type != string(workflowNodeTypeTask) {
		return fmt.Errorf("expected task type, got %s", tsk.Type)
	}
	t.TaskDef = tsk.TaskDef
	t.Type = tsk.Type
	t.ExecFn = tsk.ExecFn
	return nil
}

// UnmarshalBSON is a customer unmarshaler for taskDefParserWithoutExeFn
func (t *taskDefParserWithoutExeFn) UnmarshalBSON(data []byte) error {
	tsk := struct {
		TaskDef `bson:",inline"`
		Type    string `bson:"type"`
		ExecFn  string `bson:"execFn"`
	}{}
	err := bson.Unmarshal(data, &tsk)
	if err != nil {
		return err
	}
	if tsk.Type != string(workflowNodeTypeTask) {
		return fmt.Errorf("expected task type, got %s", tsk.Type)
	}
	t.TaskDef = tsk.TaskDef
	t.Type = tsk.Type
	t.ExecFn = tsk.ExecFn
	return nil
}

// Concrete implementation for parsing worflow defnition
// workflow are created from the instance of the workflow def
type workflowDefParser struct {
	WorkflowDef `json:",inline" bson:",inline"`
	Nodes       []workflowNodeParser `json:"nodes" bson:"nodes"`
	Type        string               `json:"type" bson:"type"`
}

func (w *workflowDefParser) UnmarshalJSON(data []byte) error {
	wrk := struct {
		WorkflowDef `json:",inline"`
		Nodes       []workflowNodeParser `json:"nodes"`
		Type        string               `json:"type"`
	}{}
	err := json.Unmarshal(data, &wrk)
	if err != nil {
		return err
	}
	if wrk.Type != string(workflowNodeTypeWorkflow) {
		return fmt.Errorf("expected workflow type, got %s", wrk.Type)
	}
	w.WorkflowDef = wrk.WorkflowDef
	w.Type = wrk.Type
	w.Nodes = wrk.Nodes
	return nil
}

func (w *workflowDefParser) UnmarshalBSON(data []byte) error {
	wrk := struct {
		WorkflowDef `bson:",inline"`
		Nodes       []workflowNodeParser `bson:"nodes"`
		Type        string               `bson:"type"`
	}{}
	err := bson.Unmarshal(data, &wrk)
	if err != nil {
		return err
	}
	if wrk.Type != string(workflowNodeTypeWorkflow) {
		return fmt.Errorf("expected workflow type, got %s", wrk.Type)
	}
	w.WorkflowDef = wrk.WorkflowDef
	w.Type = wrk.Type
	w.Nodes = wrk.Nodes
	return nil
}

type Value interface{}

// workflowNodeParser is a wrapper for workflowNode
type workflowNodeParser struct {
	Value `json:"" bson:",inline"`
	Type  workflowNodeType `json:"type" bson:"type"`
}

func (w workflowNodeParser) MarshalJSON() ([]byte, error) {
	if w.Type == workflowNodeTypeTask {
		return json.Marshal(w.Value)
	}
	if w.Type == workflowNodeTypeWorkflow {
		return json.Marshal(w.Value)
	}
	return nil, fmt.Errorf("%w, got %s", ErrUnsupportedWorkflowNode, w.Type)
}

// UnmarshalJSON is a custom unmarshaler for workflowNodeParser
// underlying type is determined by the type field and the json is unmarshaled accordingly with its concreate implementation
// support for any new workflow node type should be added here
func (w *workflowNodeParser) UnmarshalJSON(data []byte) error {
	// ew first unmarshal tom understand the type
	wrkflNode := struct {
		Type workflowNodeType `json:"type"`
	}{}
	err := json.Unmarshal(data, &wrkflNode)
	if err != nil {
		return err
	}
	if wrkflNode.Type == workflowNodeTypeTask {
		// then we unmarshal to the concrete type - task
		ts := &taskDefParserWithoutExeFn{}
		err := ts.UnmarshalJSON(data)
		if err != nil {
			return err
		}
		w.Value = ts
		w.Type = workflowNodeTypeTask
		return nil
	}
	if wrkflNode.Type == workflowNodeTypeWorkflow {
		// then we unmarshal to the concrete type - workflow
		wf := &workflowDefParser{}
		err := wf.UnmarshalJSON(data)
		if err != nil {
			return err
		}
		w.Value = wf
		w.Type = workflowNodeTypeWorkflow
		return nil
	}
	return fmt.Errorf("%w, got %s", ErrUnsupportedWorkflowNode, wrkflNode.Type)
}

func (w *workflowNodeParser) UnmarshalBSON(data []byte) error {
	// ew first unmarshal to understand the type
	wrkflNode := struct {
		Type workflowNodeType `bson:"type"`
	}{}
	err := bson.Unmarshal(data, &wrkflNode)
	if err != nil {
		return err
	}
	if wrkflNode.Type == workflowNodeTypeTask {
		// then we unmarshal to the concrete type - task
		ts := &taskDefParserWithoutExeFn{}
		err := ts.UnmarshalBSON(data)
		if err != nil {
			return err
		}
		w.Value = ts
		w.Type = workflowNodeTypeTask
		return nil
	}
	if wrkflNode.Type == workflowNodeTypeWorkflow {
		// then we unmarshal to the concrete type - workflow
		wf := &workflowDefParser{}
		err := wf.UnmarshalBSON(data)
		if err != nil {
			return err
		}
		w.Value = wf
		w.Type = workflowNodeTypeWorkflow
		return nil
	}
	return fmt.Errorf("%w, got %s", ErrUnsupportedWorkflowNode, wrkflNode.Type)
}

// GetWorkflowNode creates a workflow node from the workflowNodeParser
// It takes care of the underlying type and creates the workflow node accordingly
// It also takes care of the nested workflow nodes too
// Any new workflow node type should be added here
func (w workflowNodeParser) GetWorkflowNode(exeFns map[string]ExecutionFn) (WorkflowNode, error) {
	// based on the type we create the workflow node
	if w.Type == workflowNodeTypeTask {
		// since task is a simple workflow node, we can create it directly
		taskDef, ok := w.Value.(*taskDefParserWithoutExeFn)
		if !ok {
			return nil, fmt.Errorf("expecting a task def while parsing. got %s", reflect.TypeOf(w.Value))
		}

		exeFn, ok := exeFns[taskDef.ExecFn]
		if !ok && len(taskDef.ExecFn) > 0 {
			return nil, fmt.Errorf("couldn't find execfn %s for %s", taskDef.Name, taskDef.ExecFn)
		}
		return NewTaskDef(taskDef.Identity, taskDef.Input, taskDef.Output, exeFn).CreateTask(), nil
	} else if w.Type == workflowNodeTypeWorkflow {
		// since workflow is a complex workflow node, we need to create it recursively
		// there could be nested workflow nodes too
		wrkflwDef, ok := w.Value.(*workflowDefParser)
		if !ok {
			return nil, fmt.Errorf("expecting a workflow def while parsing. got %s", reflect.TypeOf(w.Value))
		}
		nodes := make([]WorkflowNode, len(wrkflwDef.Nodes))
		for i, node := range wrkflwDef.Nodes {
			// a recursive call to create the workflow node to handle nested workflow nodes
			wNode, err := node.GetWorkflowNode(exeFns)
			if err != nil {
				return nil, fmt.Errorf("error getting workflow node. %w", err)
			}
			nodes[i] = wNode
		}
		return NewWorkflowDef(wrkflwDef.Identity, wrkflwDef.InitialInput, wrkflwDef.OnErrorWorkFlow, nodes...).CreateWorkflow(), nil
	}
	return nil, fmt.Errorf("%w, got %s", ErrUnsupportedWorkflowNode, w.Type)
}

type WorkflowNodesDef []workflowNodeParser

// NewWorkflowNodeDefs creates a slice of workflow nodes from the parser
func NewWorkflowNodeDefs[P Parser[WorkflowNodesDef]](parser P, exeFns map[string]ExecutionFn) ([]WorkflowNode, error) {
	workflowNodeDefsP, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("error parsing workflow node def. %w", err)
	}
	if workflowNodeDefsP == nil {
		return nil, fmt.Errorf("workflow node defs is nil")
	}
	workfowNodeDefVs := *workflowNodeDefsP
	result := make([]WorkflowNode, len(workfowNodeDefVs))
	for i, workfowNodeDef := range workfowNodeDefVs {
		// we get the nodes from the node defnitions here
		workflowNode, err := workfowNodeDef.GetWorkflowNode(exeFns)
		if err != nil {
			return nil, fmt.Errorf("error getting workflow node at index %d. %w", i, err)
		}
		result[i] = workflowNode
	}

	return result, nil
}
