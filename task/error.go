package task

import "errors"

var (
	ErrInvalidDataType         = errors.New("invalid data type")
	ErrExpectingInput          = errors.New("expecting an input")
	ErrExpectingOutput         = errors.New("expecting an output")
	ErrUnsupportedWorkflowNode = errors.New("unsupported workflow node")
)
