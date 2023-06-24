package task_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/task-orc/task-core/task"
)

func TestWorkflow(t *testing.T) {

	t.Run("Simple Workflow Test", simpleWorkflowTest)

}

type simpleWorkflowTestcase struct {
	index          int
	name           string
	description    string
	testdataFile   string
	initialInput   *task.DataValue
	expectedError  error
	shouldNotParse bool
}

var simpleWorkflowTestcases = []simpleWorkflowTestcase{
	{
		index:         1,
		name:          "Simple Workflow Test for OTP",
		description:   "This is a simple workflow test for OTP",
		testdataFile:  "testdata/workflow/simple_workflow/1.json",
		initialInput:  phoneNumberInput,
		expectedError: nil,
	},
	{
		index:         2,
		name:          "Simple Workflow Test for OTP with data validation error",
		description:   "This is a simple workflow test where completion doesn't happen due to data validation error",
		testdataFile:  "testdata/workflow/simple_workflow/2.json",
		initialInput:  phoneNumberInput,
		expectedError: task.ErrInvalidDataType,
	},
	{
		index:         3,
		name:          "Simple Workflow Test for OTP with task and a workflow",
		description:   "This is a simple workflow test where we have a combination of a task and workflow",
		testdataFile:  "testdata/workflow/simple_workflow/3.json",
		initialInput:  phoneNumberInput,
		expectedError: nil,
	},
	{
		index:         4,
		name:          "Simple Workflow Test for OTP with nested workflows",
		description:   "This is a simple workflow test where we have a nested workflow",
		testdataFile:  "testdata/workflow/simple_workflow/4.json",
		initialInput:  phoneNumberInput,
		expectedError: nil,
	},
	{
		index:         5,
		name:          "Simple Workflow Test for OTP with output data validation error",
		description:   "This is a simple workflow test where completion doesn't happen due to output data validation error",
		testdataFile:  "testdata/workflow/simple_workflow/5.json",
		initialInput:  phoneNumberInput,
		expectedError: task.ErrInvalidDataType,
	},
	{
		index:         6,
		name:          "Simple Workflow Test for OTP with missing output",
		description:   "This is a simple workflow test where completion doesn't happen due to missing output",
		testdataFile:  "testdata/workflow/simple_workflow/6.json",
		initialInput:  phoneNumberInput,
		expectedError: task.ErrExpectingOutput,
	},
	{
		index:         7,
		name:          "Simple Workflow Test for OTP with missing input",
		description:   "This is a simple workflow test where completion doesn't happen due to missing input",
		testdataFile:  "testdata/workflow/simple_workflow/7.json",
		initialInput:  phoneNumberInput,
		expectedError: task.ErrExpectingInput,
	},
	{
		index:          8,
		name:           "Simple Workflow Test for OTP with unsupported node",
		description:    "This is a simple workflow test where completion doesn't happen due to unsupported node",
		testdataFile:   "testdata/workflow/simple_workflow/8.json",
		initialInput:   phoneNumberInput,
		expectedError:  task.ErrUnsupportedWorkflowNode,
		shouldNotParse: true,
	},
	{
		index:          9,
		name:           "Simple Workflow Test for OTP with inner workflow and unsupported node",
		description:    "This is a simple workflow test where completion doesn't happen due to unsupported node",
		testdataFile:   "testdata/workflow/simple_workflow/9.json",
		initialInput:   phoneNumberInput,
		expectedError:  task.ErrUnsupportedWorkflowNode,
		shouldNotParse: true,
	},
	{
		index:          10,
		name:           "Simple Workflow Test for OTP with parsing issues for task",
		description:    "This is a simple workflow test where completion doesn't happen due to parsing issue",
		testdataFile:   "testdata/workflow/simple_workflow/10.json",
		initialInput:   phoneNumberInput,
		expectedError:  errors.New("parsing"),
		shouldNotParse: true,
	},
	{
		index:          11,
		name:           "Simple Workflow Test for OTP with parsing issues for workflow",
		description:    "This is a simple workflow test where completion doesn't happen due to parsing issue",
		testdataFile:   "testdata/workflow/simple_workflow/11.json",
		initialInput:   phoneNumberInput,
		expectedError:  errors.New("parsing"),
		shouldNotParse: true,
	},
	{
		index:         12,
		name:          "Simple Workflow Test for OTP with input/output data validation issues for float",
		description:   "This is a simple workflow test where completion doesn't happen due to input/output data validation issue associated with float",
		testdataFile:  "testdata/workflow/simple_workflow/12.json",
		initialInput:  phoneNumberInput,
		expectedError: task.ErrInvalidDataType,
	},
	{
		index:         13,
		name:          "Simple Workflow Test for OTP with input/output data validation issues for int",
		description:   "This is a simple workflow test where completion doesn't happen due to input/output data validation issue associated with int",
		testdataFile:  "testdata/workflow/simple_workflow/13.json",
		initialInput:  phoneNumberInput,
		expectedError: task.ErrInvalidDataType,
	},
	{
		index:         14,
		name:          "Simple Workflow Test for OTP with input/output data validation issues for bool",
		description:   "This is a simple workflow test where completion doesn't happen due to input/output data validation issue associated with bool",
		testdataFile:  "testdata/workflow/simple_workflow/14.json",
		initialInput:  phoneNumberInput,
		expectedError: task.ErrInvalidDataType,
	},
	{
		index:         15,
		name:          "Simple Workflow Test for OTP with input/output data validation issues for string",
		description:   "This is a simple workflow test where completion doesn't happen due to input/output data validation issue associated with string",
		testdataFile:  "testdata/workflow/simple_workflow/15.json",
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
			// if testcase.index != 1 {
			// 	t.Skip()
			// }
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
				t.Errorf("error opening the testdata file for testcase %s. %s", testcase.name, err)
				return
			}
			defer f.Close()

			// create the workflow
			// we will use json parser for getting the task definitions
			parser := task.NewJsonParser[task.WorkflowNodesDef](f)
			siFns := newSimpleAsyncFns(t, ch)
			wrkflwNodes, err := task.NewWorkflowNodeDefs(parser, siFns.getSimpleAsyncFunctions())
			if err != nil && !testcase.shouldNotParse {
				t.Errorf("error parsing the testdata file for testcase %s. %s", testcase.name, err)
				return
			}
			if err == nil && testcase.shouldNotParse {
				t.Errorf("expected error while parsing the testdata file for testcase %s. got nil", testcase.name)
				return
			}
			if err != nil && testcase.shouldNotParse {
				if !strings.Contains(err.Error(), testcase.expectedError.Error()) {
					t.Errorf("expected error %s while parsing the testdata file for testcase %s. got %s", testcase.expectedError.Error(), testcase.name, err.Error())
					return
				}
				return
			}
			//finally create the workflow
			wrkFlow := task.NewWorkflowDef(task.Identity{
				Name:        testcase.name,
				Description: testcase.description,
			}, testcase.initialInput, nil, wrkflwNodes...).CreateWorkflow()
			siFns.updateWorkflowId(wrkFlow.ID)
			t.Logf("parsed and executing the workflow %s", wrkFlow.Name)

			// execute the workflow
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			doneCh := wrkflowExe.run(wrkFlow)
			defer cancel()
			select {
			case <-ctx.Done():
				break
			case <-doneCh:
				break
			}
			rpt := wrkFlow.Status()
			if rpt.Error != nil && testcase.expectedError == nil {
				t.Errorf("error executing the workflow %s. %s", wrkFlow.Name, rpt.Error.Error())
				return
			}
			if rpt.Error == nil && testcase.expectedError != nil {
				t.Errorf("error executing the workflow %s, expected error %s but got nil", wrkFlow.Name, testcase.expectedError.Error())
				return
			}
			if rpt.Error != nil && testcase.expectedError != nil && !errors.Is(rpt.Error, testcase.expectedError) {
				t.Errorf("error executing the workflow %s, expected error %s but got %s", wrkFlow.Name, testcase.expectedError.Error(), rpt.Error.Error())
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

func (w *workflowExecutioner) listen(stop <-chan bool) chan<- executionDataForWorkflow {
	ch := make(chan executionDataForWorkflow)
	go func() {
		for {
			select {
			case <-stop:
				return
			case data := <-ch:
				wrkFlow, ok := w.wrkFlows[data.workflowId]
				if !ok {
					continue
				}
				wrkFlow.UpdateStatus(data.execData)
			}
		}
	}()
	return ch
}

func (w *workflowExecutioner) run(wrkFlow *task.Workflow) chan struct{} {
	w.Lock()
	w.wrkFlows[wrkFlow.ID] = wrkFlow
	w.Unlock()
	ch := make(chan struct{})
	go func() {
		rpt := wrkFlow.Execute(nil)
		for !rpt.HasFinished {
			time.Sleep(time.Millisecond * 2)
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
	updateChan chan<- executionDataForWorkflow
	t          *testing.T
	sync.Mutex
}

type executionDataForWorkflow struct {
	workflowId string
	execData   task.ExecutionData
}

func newSimpleAsyncFns(t *testing.T, ch chan<- executionDataForWorkflow) *simpleAsyncFns {
	return &simpleAsyncFns{updateChan: ch, t: t}
}

func (s *simpleAsyncFns) updateWorkflowId(wrkFlowId string) {
	s.Lock()
	defer s.Unlock()
	s.wrkFlowId = wrkFlowId
}

func (s *simpleAsyncFns) getSimpleAsyncFunctions() map[string]task.ExecutionFn {
	return map[string]task.ExecutionFn{
		"send_otp":                task.ExecutionFn(s.dummySendOtp),
		"send_otp_without_output": task.ExecutionFn(s.dummySendOtpWithoutOutput),
		"verify_otp":              task.ExecutionFn(s.dummyVerifyOtp),
		"login_or_register_user":  task.ExecutionFn(s.dummyLoginOrRegisterUser),
	}
}

func (s *simpleAsyncFns) dummySendOtp(nodeID string, input *task.DataValue) task.ExecutionData {
	data := input.Value.(map[string]interface{})
	s.t.Logf("sending otp to %s", data["phoneNumber"])
	dt := task.ExecutionData{
		NodeId: nodeID,
		Output: &task.DataValue{
			Value: map[string]interface{}{
				"phoneNumber": data["phoneNumber"],
			},
		},
	}
	time.Sleep(time.Millisecond * 100)
	s.t.Logf("sent otp 1234 to %s", data["phoneNumber"])
	go func() {
		s.updateChan <- executionDataForWorkflow{s.wrkFlowId, dt}
	}()
	return dt
}

func (s *simpleAsyncFns) dummySendOtpWithoutOutput(nodeID string, input *task.DataValue) task.ExecutionData {
	data := input.Value.(map[string]interface{})
	s.t.Logf("sending otp to %s", data["phoneNumber"])
	dt := task.ExecutionData{
		NodeId: nodeID,
	}
	time.Sleep(time.Millisecond * 100)
	s.t.Logf("sent otp 1234 to %s", data["phoneNumber"])
	go func() {
		s.updateChan <- executionDataForWorkflow{s.wrkFlowId, dt}
	}()
	return dt
}

func (s *simpleAsyncFns) dummyVerifyOtp(nodeID string, input *task.DataValue) task.ExecutionData {
	data := input.Value.(map[string]interface{})
	s.t.Logf("verifying otp send to %s", data["phoneNumber"])
	dt := task.ExecutionData{
		NodeId: nodeID,
		Output: &task.DataValue{
			Value: map[string]interface{}{
				"phoneNumber": data["phoneNumber"],
				"verified":    true,
				"dummy_array": []interface{}{
					"1",
					"2",
				},
				"dummy_object": map[string]interface{}{
					"dummyField": "dummy_value",
				},
			},
		},
	}
	time.Sleep(time.Millisecond * 100)
	s.t.Logf("verified otp sent to %s", data["phoneNumber"])
	go func() {
		s.updateChan <- executionDataForWorkflow{s.wrkFlowId, dt}
	}()
	return dt
}

func (s *simpleAsyncFns) dummyLoginOrRegisterUser(nodeID string, input *task.DataValue) task.ExecutionData {
	data := input.Value.(map[string]interface{})
	s.t.Logf("logging in the user to %s", data["phoneNumber"])
	dt := task.ExecutionData{
		NodeId: nodeID,
	}
	time.Sleep(time.Millisecond * 100)
	s.t.Logf("successfully logged in the user %s", data["phoneNumber"])
	go func() {
		s.updateChan <- executionDataForWorkflow{s.wrkFlowId, dt}
	}()
	return dt
}
