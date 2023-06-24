package task

import (
	"fmt"
	"sync"
)

// TaskDef is a struct that represents the task definition
// Eg: {"id": "task1", "name": "Otp Verification", "description": "Verify the otp", "input": { "otp": "string" }, "output": { "isVerified": "bool" } }
// Def would be:
//
//	TaskDef{
//		ID:          "task1",
//		Name:        "Otp Verification",
//		Description: "Verify the otp",
//		Input: DataObjectDef{
//			Fields: []DataField{
//				{
//					Field: "otp",
//					Type: DataType{
//						Type: "string",
//					},
//				},
//			},
//		},
//		Output: DataObjectDef{
//			Fields: []DataField{
//				{
//					Field: "isVerified",
//					Type: DataType{
//						Type: "bool",
//					},
//				},
//			},
//		},
//	}
type TaskDef struct {
	Identity `json:",inline"`
	execFn   ExecutionFn    `json:"-"`
	Input    *DataObjectDef `json:"input"`
	Output   *DataObjectDef `json:"output"`
}

func NewTaskDef(identity Identity, input, output *DataObjectDef, execFn ExecutionFn) *TaskDef {
	taskDef := TaskDef{
		Identity: identity,
		Input:    input,
		Output:   output,
		execFn:   execFn,
	}
	if len(identity.ID) == 0 {
		taskDef.Identity = taskDef.GenerateID("taskDef_")
	}

	return &taskDef
}

func (t TaskDef) CreateTask() *Task {
	result := &Task{
		Identity:    t.Identity.GenerateID("task_"),
		HasStarted:  false,
		HasFinished: false,
		execFn:      t.execFn,
	}
	if t.Input != nil {
		result.Input = t.Input.CreateDataValue()
	}
	if t.Output != nil {
		result.Output = t.Output.CreateDataValue()
	}
	return result
}

type Task struct {
	Identity
	HasStarted  bool
	HasFinished bool
	execFn      ExecutionFn
	Input       *DataValue
	Output      *DataValue
	Error       error
	sync.Mutex
}

func (t *Task) Status() ExecutionReport {
	return ExecutionReport{
		HasStarted:  t.HasStarted,
		HasFinished: t.HasFinished,
		Input:       t.Input,
		ExecutionData: ExecutionData{
			Output: t.Output,
			Error:  t.Error,
			NodeId: t.ID,
		},
	}
}

func (t *Task) Execute(input *DataValue) ExecutionReport {
	if t.HasStarted {
		return t.Status()
	}
	t.Lock()
	t.HasStarted = true
	if t.Input != nil && (input == nil || input.Value == nil) {
		t.Error = ErrExpectingInput
		t.HasFinished = true
		t.Unlock()
		return t.Status()
	}
	if t.Input != nil {
		err := t.Input.Validate(input.Value)
		if err != nil {
			t.Error = fmt.Errorf("given input didn't match the expected input definition. %w", err)
			t.HasFinished = true
		}
		t.Input.Value = input.Value
	}
	t.Unlock()
	if t.Error != nil {
		return t.Status()
	}
	if t.execFn != nil {
		t.execFn(t.ID, input)
	}
	return t.Status()
}

func (t *Task) UpdateStatus(update ExecutionData) {
	if t.HasFinished {
		return
	}
	t.Lock()
	if t.Output != nil && update.Error == nil && (update.Output == nil || update.Output.Value == nil) {
		t.Error = ErrExpectingOutput
		t.HasFinished = true
		t.Unlock()
		return
	}
	if t.Output != nil && update.Error == nil {
		err := t.Output.Validate(update.Output.Value)
		if err != nil {
			t.Error = fmt.Errorf("given output didn't match the expected output definition. %w", err)
			t.HasFinished = true
		}
		t.Output.Value = update.Output.Value
	}
	t.Error = update.Error
	t.HasFinished = true
	t.Unlock()
}
