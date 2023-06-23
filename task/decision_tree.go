package task

type DecisionTree struct {
	Identity
	PossibleSteps []DecisionTreeStep
}

type DecisionTreeStep struct {
	Condition string
	Step      Workflow
}
