package swf

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

//ComposedDecider can be used to build a decider out of a number of sub Deciders
//the sub deciders should return Pass when they dont wish to handle an event.
type ComposedDecider struct {
	deciders []Decider
	last     bool
}

//NewComposedDecider builds a Composed Decider from a list of sub Deciders.
//You can compose your fiinal composable decider from other composable deciders,
//but you should make sure that the final decider includes a 'catch-all' decider in last place
//you can use DefaultDecider() or your own.
func NewComposedDecider(deciders ...Decider) Decider {
	c := ComposedDecider{
		deciders: deciders,
	}
	return c.Decide
}

//Decide is the the Decider func for a ComposedDecider
func (c *ComposedDecider) Decide(ctx *FSMContext, h HistoryEvent, data interface{}) Outcome {
	decisions := ctx.EmptyDecisions()
	for _, d := range c.deciders {
		outcome := d(ctx, h, data)
		if outcome == Pass {
			continue
		}

		// contribute the outcome's decisions and data
		decisions = append(decisions, outcome.Decisions()...)
		data = outcome.Data()
		switch outcome.(type) {
		case ContinueOutcome:
			// ContinueOutcome's only job is to contribute to later outcomes
			continue
		default:
			return TransitionOutcome{
				data:      data,
				state:     outcome.State(),
				decisions: decisions,
			}
		}
	}
	return ContinueOutcome{
		data:      data,
		decisions: decisions,
	}
}

//DefaultDecider is a 'catch-all' decider that simply logs the unhandled decision.
//You should place this or one like it as the last decider in your top level ComposableDecider.
func DefaultDecider() Decider {
	return func(ctx *FSMContext, h HistoryEvent, data interface{}) Outcome {
		log.Printf("at=unhandled-event event=%s state=%s default=stay decisions=0", h.EventType, ctx.State)
		return ctx.Stay(data, ctx.EmptyDecisions())
	}
}

//DecisionFunc is a building block for composable deciders that returns a decision.
type DecisionFunc func(ctx *FSMContext, h HistoryEvent, data interface{}) Decision

//MultiDecisionFunc is a building block for composable deciders that returns a [] of decision.
type MultiDecisionFunc func(ctx *FSMContext, h HistoryEvent, data interface{}) []Decision

//StateFunc is a building block for composable deciders mutates the FSM stateData.
type StateFunc func(ctx *FSMContext, h HistoryEvent, data interface{})

//PredicateFunc is a building block for composable deciders, a predicate based on the FSM stateData.
type PredicateFunc func(data interface{}) bool

//Typed allows you to create Typed building blocks for composable deciders.
//the type checking here is done on constriction at runtime, so be sure to have a unit test that constructs your funcs.
func Typed(typed interface{}) *TypedFuncs {
	return &TypedFuncs{typed}
}

//TypedFuncs lets you construct building block for composable deciders, that have arguments that are checked
//against the type of your FSM stateData.
type TypedFuncs struct {
	typed interface{}
}

func (t *TypedFuncs) typeArg() string {
	return reflect.TypeOf(t.typed).String()
}

//Decider builds a Decider from your typed Decider that verifies the right typing at construction time.
func (t *TypedFuncs) Decider(decider interface{}) Decider {
	typeCheck(decider, []string{"*swf.FSMContext", "swf.HistoryEvent", t.typeArg()}, []string{"swf.Outcome"})
	return marshalledFunc{reflect.ValueOf(decider)}.decider
}

//DecisionFunc builds a DecisionFunc from your typed DecisionFunc that verifies the right typing at construction time.
func (t *TypedFuncs) DecisionFunc(decisionFunc interface{}) DecisionFunc {
	typeCheck(decisionFunc, []string{"*swf.FSMContext", "swf.HistoryEvent", t.typeArg()}, []string{"swf.Decision"})
	return marshalledFunc{reflect.ValueOf(decisionFunc)}.decisionFunc
}

//MultiDecisionFunc builds a MultiDecisionFunc from your typed MultiDecisionFunc that verifies the right typing at construction time.
func (t *TypedFuncs) MultiDecisionFunc(decisionFunc interface{}) MultiDecisionFunc {
	typeCheck(decisionFunc, []string{"*swf.FSMContext", "swf.HistoryEvent", t.typeArg()}, []string{"[]swf.Decision"})
	return marshalledFunc{reflect.ValueOf(decisionFunc)}.multiDecisionFunc
}

//StateFunc builds a StateFunc from your typed StateFunc that verifies the right typing at construction time.
func (t *TypedFuncs) StateFunc(stateFunc interface{}) StateFunc {
	typeCheck(stateFunc, []string{"*swf.FSMContext", "swf.HistoryEvent", t.typeArg()}, []string{})
	return marshalledFunc{reflect.ValueOf(stateFunc)}.stateFunc
}

//PredicateFunc builds a PredicateFunc from your typed PredicateFunc that verifies the right typing at construction time.
func (t *TypedFuncs) PredicateFunc(stateFunc interface{}) PredicateFunc {
	typeCheck(stateFunc, []string{t.typeArg()}, []string{"bool"})
	return marshalledFunc{reflect.ValueOf(stateFunc)}.predicateFunc
}

type marshalledFunc struct {
	v reflect.Value
}

func (m marshalledFunc) decider(f *FSMContext, h HistoryEvent, data interface{}) Outcome {
	ret := m.v.Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(h), reflect.ValueOf(data)})[0]
	if ret.IsNil() {
		return nil
	}
	return ret.Interface().(Outcome)
}

func (m marshalledFunc) decisionFunc(f *FSMContext, h HistoryEvent, data interface{}) Decision {
	ret := m.v.Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(h), reflect.ValueOf(data)})[0]
	return ret.Interface().(Decision)
}

func (m marshalledFunc) multiDecisionFunc(f *FSMContext, h HistoryEvent, data interface{}) []Decision {
	ret := m.v.Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(h), reflect.ValueOf(data)})[0]
	return ret.Interface().([]Decision)
}

func (m marshalledFunc) stateFunc(f *FSMContext, h HistoryEvent, data interface{}) {
	m.v.Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(h), reflect.ValueOf(data)})
}

func (m marshalledFunc) predicateFunc(data interface{}) bool {
	return m.v.Call([]reflect.Value{reflect.ValueOf(data)})[0].Interface().(bool)
}

func typeCheck(typedFunc interface{}, in []string, out []string) {
	t := reflect.TypeOf(typedFunc)
	if reflect.Func != t.Kind() {
		panic(fmt.Sprintf("kind was %v, not Func", t.Kind()))
	}
	if len(in) != t.NumIn() {
		panic(fmt.Sprintf(
			"input arity was %v, not %v",
			t.NumIn(), len(in),
		))
	}

	for i, rt := range in {
		if rt != t.In(i).String() {
			panic(fmt.Sprintf(
				"type of argument %v was %v, not %v",
				i, t.In(i), rt,
			))
		}
	}

	if len(out) != t.NumOut() {
		panic(fmt.Sprintf(
			"number of return values was %v, not %v",
			t.NumOut(), len(out),
		))
	}

	for i, rt := range out {
		if rt != t.Out(i).String() {
			panic(fmt.Sprintf(
				"type of return value %v was %v, not %v",
				i, t.Out(i), rt,
			))
		}
	}
}

//ManagedContinuations is a composable decider that will handle most of the mechanics of autmoatically continuing workflows.
//todo, does it ever happen that we would have a decision task that previous deciders would have created a decision that
//breaks continuation? i.e. you get decision task that has a history containing 2 signals, etc.
//lets assume not till we find otherwise. If we are wrong, managedcontinuations probably cant happen in userspace.
// If there are no activities present in the tracker, it will continueAsNew the workflow in response
// to a FSM.ContinueWorkflow timer or signal. If there are activities present in the tracker, it will
// set a new FSM.ContinueWorkflow timer, that fires in timerRetrySeconds.
// It will also signal the workflow to continue when the workflow history grows beyond the
// configured historySize.
// this should be last in your decider stack, as it will signal in response to *any* event that
// has an id > historySize
func ManagedContinuations(historySize int, timerRetrySeconds int) Decider {
	handleContinuationTimer := func(ctx *FSMContext, h HistoryEvent, data interface{}) Outcome {
		if h.EventType == EventTypeTimerFired && h.TimerFiredEventAttributes.TimerID == ContinueTimer {
			if len(ctx.ActivitiesInfo()) == 0 {
				decisions := append(ctx.EmptyDecisions(), ctx.ContinueWorkflowDecision(ctx.State))
				return ctx.Stay(data, decisions)
			}
			d := Decision{
				DecisionType: DecisionTypeStartTimer,
				StartTimerDecisionAttributes: &StartTimerDecisionAttributes{
					StartToFireTimeout: strconv.Itoa(timerRetrySeconds),
					TimerID:            ContinueTimer,
				},
			}
			decisions := append(ctx.EmptyDecisions(), d)
			return ctx.Stay(data, decisions)

		}
		return Pass
	}

	handleContinuationSignal := func(ctx *FSMContext, h HistoryEvent, data interface{}) Outcome {
		if h.EventType == EventTypeWorkflowExecutionSignaled && h.WorkflowExecutionSignaledEventAttributes.SignalName == ContinueSignal {

			if len(ctx.ActivitiesInfo()) == 0 {
				decisions := append(ctx.EmptyDecisions(), ctx.ContinueWorkflowDecision(ctx.State))
				return ctx.Stay(data, decisions)
			}

			d := Decision{
				DecisionType: DecisionTypeStartTimer,
				StartTimerDecisionAttributes: &StartTimerDecisionAttributes{
					StartToFireTimeout: strconv.Itoa(timerRetrySeconds),
					TimerID:            ContinueTimer,
				},
			}
			decisions := append(ctx.EmptyDecisions(), d)
			return ctx.Stay(data, decisions)

		}
		return Pass
	}

	signalContinuationWhenHistoryLarge := func(ctx *FSMContext, h HistoryEvent, data interface{}) Outcome {
		if h.EventID > historySize {
			d := Decision{
				DecisionType: DecisionTypeSignalExternalWorkflowExecution,
				SignalExternalWorkflowExecutionDecisionAttributes: &SignalExternalWorkflowExecutionDecisionAttributes{
					SignalName: ContinueSignal,
					WorkflowID: ctx.WorkflowID,
					RunID:      ctx.RunID,
				},
			}
			decisions := append(ctx.EmptyDecisions(), d)
			return ctx.Stay(data, decisions)
		}
		return Pass
	}

	return NewComposedDecider(
		handleContinuationTimer,
		handleContinuationSignal,
		signalContinuationWhenHistoryLarge,
	)

}
