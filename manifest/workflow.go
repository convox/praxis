package manifest

type Workflow struct {
	Type    string
	Trigger string
	Steps   WorkflowSteps
}

type Workflows []Workflow

type WorkflowStep struct {
	Type   string
	Target string
}

type WorkflowSteps []WorkflowStep

func (w *Workflows) Find(typ, trigger string) *Workflow {
	for _, wf := range *w {
		if wf.Type == typ && wf.Trigger == trigger {
			return &wf
		}
	}

	return nil
}
