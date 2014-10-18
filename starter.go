package swf

type Starter struct {
	client WorkflowClient
}

func (s *Starter) run() {
	s.client.StartWorkflow(
		StartWorkflowRequest{
			Domain: "initialTest",
			WorkflowType: WorkflowType{
				Name:    "initialTest",
				Version: "1",
			},
			WorkflowId: "uuid"})
}

func NewStarter() *Starter {
	return &Starter{
		client: &Client{},
	}
}
