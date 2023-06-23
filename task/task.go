package task

import "sync"

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
	Identity
	ExecFn func(input *DataValue) ExecutionData
	Input  *DataObjectDef
	Output *DataObjectDef
}

func (t TaskDef) CreateTask() *Task {
	result := &Task{
		Identity:    t.Identity.GenerateID("task_"),
		HasStarted:  false,
		HasFinished: false,
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
	ExecFn      func(input *DataValue) ExecutionData
	Input       *DataValue
	Output      *DataValue
	Error       error
	sync.Mutex
}

func (t *Task) Status() ExecutionData {
	return ExecutionData{
		HasStarted:  t.HasStarted,
		HasFinished: t.HasFinished,
		Input:       t.Input,
		Output:      t.Output,
		Error:       t.Error,
		NodeId:      t.ID,
	}
}

func (t *Task) Execute(input *DataValue) ExecutionData {
	if t.HasStarted {
		return t.Status()
	}
	t.Lock()
	t.HasStarted = true
	//TODO: Validate input of task with task definition
	// t.Input = input
	// t.Output = t.Output
	t.Unlock()
	exeData := t.ExecFn(input)
	t.Lock()
	t.Output = exeData.Output
	t.Error = exeData.Error
	t.HasFinished = exeData.HasFinished
	t.Unlock()
	return t.Status()
}
