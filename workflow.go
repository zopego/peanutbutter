package panelbubble

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Workflow steps are chained together to form a workflow
// Each step is associated with a panel/ui element, and
// it lets the user move from one step to the next for a selected workflow

// WorkflowHandler is a struct that describes a step in a workflow
// It references the workflow it belongs to
// And the handler function that implements the logic for this step
type WorkflowHandler[S any] struct {
	Workflow    *WorkFlow[S]
	step_number int
	Handler     func(model tea.Model, state S) (tea.Model, S, tea.Cmd)
}

// WorkFlow is a struct that describes a workflow
// It contains the name of the workflow
// And the current state of the workflow
type WorkFlow[S any] struct {
	name         string
	state        S
	num_handlers int
}

func NewWorkflow[S any](name string) *WorkFlow[S] {
	return &WorkFlow[S]{
		name: name,
	}
}

func (w *WorkFlow[S]) AddHandler(handler *WorkflowHandler[S]) {
	w.num_handlers++
	handler.step_number = w.num_handlers - 1
	handler.Workflow = w
}

// For panels that are a part of a workflow, the workflow handles
// checking if the focus-grant message is relevant to this step
func (w WorkflowHandler[S]) HandleFocusGrant(model tea.Model, msg FocusGrantMsg) (tea.Model, tea.Cmd) {
	// Here we need to check if
	if msg.WorkflowName == w.Workflow.name &&
		(msg.Relation == NextWorkflow || msg.Relation == PrevWorkflow) {
		var next_step int = -1
		switch msg.Relation {
		case NextWorkflow:
			next_step = msg.Path[0] + 1
		case PrevWorkflow:
			next_step = msg.Path[0] - 1
		}
		if next_step >= 0 && next_step < w.Workflow.num_handlers {
			if next_step == w.step_number {
				m, s, cmd := w.Handler(model, w.Workflow.state)
				w.Workflow.state = s
				return m, cmd
			}
		}
	}
	return model, nil
}

func (w WorkflowHandler[S]) GetNumber() int {
	return w.step_number
}

func (w WorkflowHandler[S]) GetWorkflowName() string {
	return w.Workflow.name
}

func (w WorkflowHandler[S]) IsFirst() bool {
	return w.step_number == 0
}

func (w WorkflowHandler[S]) IsLast() bool {
	return w.step_number == w.Workflow.num_handlers-1
}
