package task

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type ParseOutput interface {
	WorkflowNodesDef
}

type Parser[o ParseOutput] interface {
	Parse() (*o, error)
}

type workflowNodeType string

const (
	workflowNodeTypeTask     workflowNodeType = "task"
	workflowNodeTypeWorkflow workflowNodeType = "workflow"
)

type taskDefParserWithoutExeFn struct {
	TaskDef `json:",inline"`
	Type    string `json:"type"`
	ExecFn  string `json:"execFn"`
}

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

type workflowDefParser struct {
	WorkflowDef `json:",inline"`
	Nodes       []workflowNodeParser `json:"nodes"`
	Type        string               `json:"type"`
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

type workflowNodeParser struct {
	Value interface{}      `json:",inline"`
	Type  workflowNodeType `json:"type"`
}

func (w *workflowNodeParser) UnmarshalJSON(data []byte) error {
	wrkflNode := struct {
		Type workflowNodeType `json:"type"`
	}{}
	err := json.Unmarshal(data, &wrkflNode)
	if err != nil {
		return err
	}
	if wrkflNode.Type == workflowNodeTypeTask {
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

func (w workflowNodeParser) GetWorkflowNode(exeFns map[string]ExecutionFn) (WorkflowNode, error) {
	if w.Type == workflowNodeTypeTask {
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
		wrkflwDef, ok := w.Value.(*workflowDefParser)
		if !ok {
			return nil, fmt.Errorf("expecting a workflow def while parsing. got %s", reflect.TypeOf(w.Value))
		}
		nodes := make([]WorkflowNode, len(wrkflwDef.Nodes))
		for i, node := range wrkflwDef.Nodes {
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
		workflowNode, err := workfowNodeDef.GetWorkflowNode(exeFns)
		if err != nil {
			return nil, fmt.Errorf("error getting workflow node at index %d. %w", i, err)
		}
		result[i] = workflowNode
	}

	return result, nil
}
