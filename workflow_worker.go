package swf

type WorkflowWorker struct {
	client          *Client
	stateSerializer StateSerializer
	idGenerator     IdGenerator
}

func NewWorkflowWorker(client *Client, stateSerializer StateSerializer, idGenerator IdGenerator) *WorkflowWorker {
	return &WorkflowWorker{client: client, stateSerializer: stateSerializer, idGenerator: idGenerator}
}

func (w *WorkflowWorker) StartWorkflow(domain string, workflowName string, workflowVersion string, input interface{}) (string, error) {
	serialized, err := w.stateSerializer.Serialize(input)
	if err != nil {
		return "", err
	}
	workflowId := w.idGenerator.WorkflowID()
	_, err = w.client.StartWorkflow(
		StartWorkflowRequest{
			Domain:     domain,
			Input:      serialized,
			WorkflowId: workflowId,
			WorkflowType: WorkflowType{
				Name:    workflowName,
				Version: workflowVersion},
			TaskStartToCloseTimeout: "20",
		})
	if err != nil {
		return "", err
	}

	return workflowId, err

}

func (w *WorkflowWorker) TerminateWorkflow(domain string, workflowId string) error {
	return w.client.TerminateWorkflowExecution(TerminateWorkflowExecution{
		Domain:     domain,
		WorkflowId: workflowId,
	})
}
