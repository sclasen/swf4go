package swf

import "code.google.com/p/go-uuid/uuid"

//probably going to want to thread data through. workflowId to ActivityId()?
type IdGenerator interface {
	ActivityID() string
	WorkflowID() string
}

type UUIDGenerator struct{}

func (u UUIDGenerator) ActivityID() string {
	return uuid.New()
}

func (u UUIDGenerator) WorkflowID() string {
	return uuid.New()
}
