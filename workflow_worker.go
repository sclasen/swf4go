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
	workflowId := w.idGenerator.WorkflowID()
	return w.StartWorkflowWithId(domain, workflowName, workflowVersion, workflowId, input)
}

func (w *WorkflowWorker) StartWorkflowWithId(domain string, workflowName string, workflowVersion string, workflowId string, input interface{}) (string, error) {
	serialized, err := w.stateSerializer.Serialize(input)
	if err != nil {
		return "", err
	}

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

func (w WorkflowWorker) SignalWorkflowExecution(domain string, workflowId string, signal string, input interface{}) error {
	serialized, err := w.stateSerializer.Serialize(input)
	if err != nil {
		return err
	}
	return w.client.SignalWorkflow(SignalWorkflowRequest{Domain: domain, WorkflowId:workflowId, SignalName:signal, Input:serialized})
}
