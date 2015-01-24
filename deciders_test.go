package swf

import (
	"testing"
)

func TestTypedFuncs(t *testing.T) {
	typedFuncs := Typed(new(TestingType))
	typedFuncs.Decider(TestingDecider)(new(FSMContext), HistoryEvent{}, new(TestingType))
	typedFuncs.DecisionFunc(TestingDecisionFunc)(new(FSMContext), HistoryEvent{}, new(TestingType))
	typedFuncs.MultiDecisionFunc(TestingMultiDecisionFunc)(new(FSMContext), HistoryEvent{}, new(TestingType))
	typedFuncs.PredicateFunc(TestingPredicateFunc)(new(TestingType))
	typedFuncs.StateFunc(TestingStateFunc)(new(FSMContext), HistoryEvent{}, new(TestingType))
}

type TestingType struct {
	Field string
}

func TestingDecider(ctx *FSMContext, h HistoryEvent, data *TestingType) Outcome {
	return ctx.Stay(data, ctx.EmptyDecisions())
}

func TestingDecisionFunc(ctx *FSMContext, h HistoryEvent, data *TestingType) Decision {
	return Decision{}
}

func TestingMultiDecisionFunc(ctx *FSMContext, h HistoryEvent, data *TestingType) []Decision {
	return []Decision{Decision{}}
}

func TestingPredicateFunc(data *TestingType) bool {
	return false
}

func TestingStateFunc(ctx *FSMContext, h HistoryEvent, data *TestingType) {
}

func TestComposedDecider(t *testing.T) {
	typedFuncs := Typed(new(TestingType))
	composed := NewComposedDecider(
		typedFuncs.Decider(TestingDecider),
	)
	composed(new(FSMContext), HistoryEvent{}, new(TestingType))
}
