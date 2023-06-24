package task

import "fmt"

type ParseOutput interface {
	TaskDef | TaskDefs
}

type Parser[o ParseOutput] interface {
	Parse() (*o, error)
}

type taskDefParserWithoutExeFn struct {
	TaskDef `json:",inline"`
	ExecFn  string `json:"execFn"`
}

type TaskDefs []taskDefParserWithoutExeFn

func NewTaskDefs[P Parser[TaskDefs]](parser P, exeFns map[string]ExecutionFn) ([]TaskDef, error) {
	taskDefsP, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("error parsing task def %w", err)
	}
	if taskDefsP == nil {
		return nil, fmt.Errorf("task defs is nil")
	}
	taskDefs := *taskDefsP
	result := make([]TaskDef, len(taskDefs))
	for i, taskDef := range taskDefs {
		exeFn, ok := exeFns[taskDef.ExecFn]
		if !ok {
			return nil, fmt.Errorf("couldn't find execfn %s for %s at index %d", taskDef.Name, taskDef.ExecFn, i)
		}
		result[i] = NewTaskDef(taskDef.Identity, taskDef.Input, taskDef.Output, exeFn)
	}

	return result, nil
}
