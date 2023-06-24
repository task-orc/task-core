package task

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type ParseOutput interface {
	WorkflowNodeDef
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
	Type        string `json:"type"`
}

func (w *workflowDefParser) UnmarshalJSON(data []byte) error {
	wrk := struct {
		WorkflowDef `json:",inline"`
		Type        string `json:"type"`
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
		w.Value = ts
		w.Type = workflowNodeTypeTask
		return ts.UnmarshalJSON(data)
	}
	if wrkflNode.Type == workflowNodeTypeWorkflow {
		wf := &workflowDefParser{}
		w.Value = wf
		w.Type = workflowNodeTypeWorkflow
		return wf.UnmarshalJSON(data)
	}
	return fmt.Errorf("expected workflow node type, got %s", wrkflNode.Type)
}

type WorkflowNodeDef []workflowNodeParser

func NewWorkflowNodeDefs[P Parser[WorkflowNodeDef]](parser P, exeFns map[string]ExecutionFn) ([]WorkflowNode, error) {
	workflowNodeDefsP, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("error parsing workflow node def %w", err)
	}
	if workflowNodeDefsP == nil {
		return nil, fmt.Errorf("workflow node defs is nil")
	}
	workfowNodeDefVs := *workflowNodeDefsP
	result := make([]WorkflowNode, len(workfowNodeDefVs))
	for i, workfowNodeDef := range workfowNodeDefVs {
		if workfowNodeDef.Type == workflowNodeTypeTask {
			taskDef, ok := workfowNodeDef.Value.(*taskDefParserWithoutExeFn)
			if !ok {
				return nil, fmt.Errorf("expecting a task def while parsing. got %s", reflect.TypeOf(workfowNodeDef.Value))
			}

			exeFn, ok := exeFns[taskDef.ExecFn]
			if !ok && len(taskDef.ExecFn) > 0 {
				return nil, fmt.Errorf("couldn't find execfn %s for %s at index %d", taskDef.Name, taskDef.ExecFn, i)
			}
			result[i] = NewTaskDef(taskDef.Identity, taskDef.Input, taskDef.Output, exeFn).CreateTask()
		} else if workfowNodeDef.Type == workflowNodeTypeWorkflow {
			wrkflwDef, ok := workfowNodeDef.Value.(*workflowDefParser)
			if !ok {
				return nil, fmt.Errorf("expecting a workflow def while parsing. got %s", reflect.TypeOf(workfowNodeDef.Value))
			}
			result[i] = NewWorkflowDef(wrkflwDef.Identity, wrkflwDef.InitialInput, wrkflwDef.OnErrorWorkFlow, wrkflwDef.Nodes...).CreateWorkflow()
		} else {
			return nil, fmt.Errorf("expecting a workflow node def while parsing. got %s", reflect.TypeOf(workfowNodeDef.Value))
		}
	}

	return result, nil
}
