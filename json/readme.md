
```shell
go get github.com/bemasher/JSONGen
jsongen json/StartWorkflowExecutionRequest.json
```

```go
type _ struct {
        ChildPolicy                  string   `json:"childPolicy"`
        Domain                       string   `json:"domain"`
        ExecutionStartToCloseTimeout string   `json:"executionStartToCloseTimeout"`
        Input                        string   `json:"input"`
        TagList                      []string `json:"tagList"`
        TaskList                     struct {
                Name string `json:"name"`
        } `json:"taskList"`
        TaskStartToCloseTimeout string `json:"taskStartToCloseTimeout"`
        WorkflowId              string `json:"workflowId"`
        WorkflowType            struct {
                Name    string `json:"name"`
                Version string `json:"version"`
        } `json:"workflowType"`
}
```