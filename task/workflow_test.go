package task_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/task-orc/task-core/task"
)

func TestWorkflow(t *testing.T) {

	t.Run("Simple Workflow Test", simpleWorkflowTest)

}

type simpleWorkflowTestcase struct {
	index         int
	name          string
	description   string
	testdataFile  string
	initialInput  *task.DataValue
	expectedError error
}

var simpleWorkflowTestcases = []simpleWorkflowTestcase{
	{
		index:         0,
		name:          "Simple Workflow Test for OTP",
		description:   "This is a simple workflow test for OTP",
		testdataFile:  "testdata/workflow/simple_workflow/1.json",
		initialInput:  phoneNumberInput,
		expectedError: nil,
	},
	{
		index:         1,
		name:          "Simple Workflow Test for OTP with data validation error",
		description:   "This is a simple workflow test where completion doesn't happen due to data validation error",
		testdataFile:  "testdata/workflow/simple_workflow/2.json",
		initialInput:  phoneNumberInput,
		expectedError: task.ErrInvalidDataType,
	},
}

// we will test simple workflow with 3 task nodes
func simpleWorkflowTest(t *testing.T) {
	/**
	 * create a workflow executioner
	 * listen to the updates
	 * run the testcases
	 */
	// created the workflow executioner
	wrkflowExe := newWorkflowExecutioner()

	// listen to the updates
	stopCh := make(chan bool)
	defer close(stopCh)
	ch := wrkflowExe.listen(stopCh)

	// run the testcases
	for _, testcase := range simpleWorkflowTestcases {
		t.Run(fmt.Sprintf("%d %s", testcase.index, testcase.name), func(t *testing.T) {
			/**
			 * read the testdata
			 * create the workflow
			 * execute the workflow
			 * check the output
			 */
			t.Logf("running the test case %s", testcase.name)

			// read the testdata
			f, err := os.Open(testcase.testdataFile)
			if err != nil {
				t.Errorf("error opening the testdata file for testcase %s %s", testcase.name, err)
				return
			}
			defer f.Close()

			// create the workflow
			// we will use json parser for getting the task definitions
			parser := task.NewJsonParser[task.TaskDefs](f)
			siFns := newSimpleAsyncFns(t, ch)
			taskDefs, err := task.NewTaskDefs(parser, siFns.getSimpleAsyncFunctions())
			if err != nil {
				t.Errorf("error parsing the testdata file for testcase %s %s", testcase.name, err)
				return
			}
			// will convert the task defs to workflow nodes
			wrkflwNodes := make([]task.WorkflowNode, len(taskDefs))
			for i, taskDef := range taskDefs {
				wrkflwNodes[i] = taskDef.CreateTask()
			}
			//finally create the workflow
			wrkFlow := task.NewWorkflow(task.Identity{
				Name:        testcase.name,
				Description: testcase.description,
			}, nil, wrkflwNodes...)
			siFns.updateWorkflowId(wrkFlow.ID)
			t.Logf("parsed and executing the workflow %s", wrkFlow.Name)

			// execute the workflow
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			doneCh := wrkflowExe.run(wrkFlow, testcase.initialInput)
			defer cancel()
			select {
			case <-ctx.Done():
				break
			case <-doneCh:
				break
			}
			rpt := wrkFlow.Status()
			if rpt.Error != nil && testcase.expectedError == nil {
				t.Errorf("error executing the workflow %s %s", wrkFlow.Name, rpt.Error.Error())
				return
			}
			if rpt.Error == nil && testcase.expectedError != nil {
				t.Errorf("error executing the workflow %s expected error %s but got nil", wrkFlow.Name, testcase.expectedError.Error())
				return
			}
			if rpt.Error != nil && testcase.expectedError != nil && !errors.Is(rpt.Error, testcase.expectedError) {
				t.Errorf("error executing the workflow %s expected error %s but got %s", wrkFlow.Name, testcase.expectedError.Error(), rpt.Error.Error())
				return
			}
			if rpt.HasFinished == false {
				t.Errorf("workflow %s has not finished in the deadline time", wrkFlow.Name)
				return
			}
			t.Logf("workflow %s has finished", wrkFlow.Name)
		})
	}
}

type workflowExecutioner struct {
	wrkFlows map[string]*task.Workflow
	sync.Mutex
}

func newWorkflowExecutioner() *workflowExecutioner {
	return &workflowExecutioner{
		wrkFlows: make(map[string]*task.Workflow),
	}
}

func (w *workflowExecutioner) listen(stop <-chan bool) chan<- task.ExecutionData {
	ch := make(chan task.ExecutionData)
	go func() {
		for {
			select {
			case <-stop:
				return
			case data := <-ch:
				wrkFlow, ok := w.wrkFlows[data.NodeId]
				if !ok {
					continue
				}
				wrkFlow.UpdateStatus(data)
			}
		}
	}()
	return ch
}

func (w *workflowExecutioner) run(wrkFlow *task.Workflow, initialInput *task.DataValue) chan struct{} {
	w.Lock()
	defer w.Unlock()
	w.wrkFlows[wrkFlow.ID] = wrkFlow
	ch := make(chan struct{})
	go func() {
		rpt := wrkFlow.Execute(initialInput)
		for !rpt.HasFinished {
			err := wrkFlow.Run()
			if err != nil {
				break
			}
			rpt = wrkFlow.Status()
		}
		ch <- struct{}{}
	}()
	return ch
}

var phoneNumberInput = &task.DataValue{
	Value: map[string]interface{}{
		"phoneNumber": "1234567890",
	},
	DataType: task.DataType{
		Type: task.DataTypeObject,
		ObjectDef: &task.DataObjectDef{
			Fields: []task.DataField{
				{
					Field:      "phoneNumber",
					IsRequired: true,
					Type: task.DataType{
						Type: task.DataTypeString,
					},
				},
			},
		},
	},
}

type simpleAsyncFns struct {
	wrkFlowId  string
	updateChan chan<- task.ExecutionData
	t          *testing.T
	sync.Mutex
}

func newSimpleAsyncFns(t *testing.T, ch chan<- task.ExecutionData) *simpleAsyncFns {
	return &simpleAsyncFns{updateChan: ch, t: t}
}

func (s *simpleAsyncFns) updateWorkflowId(wrkFlowId string) {
	s.Lock()
	defer s.Unlock()
	s.wrkFlowId = wrkFlowId
}

func (s *simpleAsyncFns) getSimpleAsyncFunctions() map[string]task.ExecutionFn {
	return map[string]task.ExecutionFn{
		"send_otp":               task.ExecutionFn(s.dummySendOtp),
		"verify_otp":             task.ExecutionFn(s.dummyVerifyOtp),
		"login_or_register_user": task.ExecutionFn(s.dummyLoginOrRegisterUser),
	}
}

func (s *simpleAsyncFns) dummySendOtp(input *task.DataValue) task.ExecutionData {
	data := input.Value.(map[string]interface{})
	s.t.Logf("sending otp to %s", data["phoneNumber"])
	dt := task.ExecutionData{
		NodeId: s.wrkFlowId,
		Output: &task.DataValue{
			Value: map[string]interface{}{
				"phoneNumber": data["phoneNumber"],
			},
		},
	}
	time.Sleep(time.Second * 2)
	s.t.Logf("sent otp 1234 to %s", data["phoneNumber"])
	go func() {
		s.updateChan <- dt
	}()
	return dt
}

func (s *simpleAsyncFns) dummyVerifyOtp(input *task.DataValue) task.ExecutionData {
	data := input.Value.(map[string]interface{})
	s.t.Logf("verifying otp send to %s", data["phoneNumber"])
	dt := task.ExecutionData{
		NodeId: s.wrkFlowId,
		Output: &task.DataValue{
			Value: map[string]interface{}{
				"phoneNumber": data["phoneNumber"],
				"verified":    true,
			},
		},
	}
	time.Sleep(time.Second * 2)
	s.t.Logf("verified otp sent to %s", data["phoneNumber"])
	go func() {
		s.updateChan <- dt
	}()
	return dt
}

func (s *simpleAsyncFns) dummyLoginOrRegisterUser(input *task.DataValue) task.ExecutionData {
	data := input.Value.(map[string]interface{})
	s.t.Logf("logging in the user to %s", data["phoneNumber"])
	dt := task.ExecutionData{
		NodeId: s.wrkFlowId,
	}
	time.Sleep(time.Second * 2)
	s.t.Logf("successfully logged in the user %s", data["phoneNumber"])
	go func() {
		s.updateChan <- dt
	}()
	return dt
}
