package swf

import (
	"fmt"
	"log"
	"reflect"
)

type ComposedDecider struct {
	deciders []Decider
}

func NewComposedDecider(deciders ...Decider) Decider {
	c := ComposedDecider{
		deciders: deciders,
	}
	return c.Decide
}

func (c *ComposedDecider) Decide(ctx *FSMContext, h HistoryEvent, data interface{}) Outcome {
	for _, d := range c.deciders {
		outcome := d(ctx, h, data)
		if outcome != nil {
			return outcome
		}
	}
	log.Printf("at=unhandled-event event=%s state=%s default=stay decisions=0", h.EventType, ctx.State)
	return ctx.Stay(data, ctx.EmptyDecisions())
}

type DecisionFunc func(ctx *FSMContext, h HistoryEvent, data interface{}) Decision

type MultiDecisionFunc func(ctx *FSMContext, h HistoryEvent, data interface{}) []Decision

type StateFunc func(ctx *FSMContext, h HistoryEvent, data interface{})

type PredicateFunc func(data interface{}) bool

func Typed(typed interface{}) *TypedFuncs {
	return &TypedFuncs{typed}
}

type TypedFuncs struct {
	typed interface{}
}

func (t *TypedFuncs) typeArg() string {
	return reflect.TypeOf(t.typed).String()
}

func (t *TypedFuncs) Decider(decider interface{}) Decider {
	typeCheck(decider, []string{"*swf.FSMContext", "swf.HistoryEvent", t.typeArg()}, []string{"swf.Outcome"})
	return MarshalledFunc{reflect.ValueOf(decider)}.Decider
}

func (t *TypedFuncs) DecisionFunc(decisionFunc interface{}) DecisionFunc {
	typeCheck(decisionFunc, []string{"*swf.FSMContext", "swf.HistoryEvent", t.typeArg()}, []string{"swf.Decision"})
	return MarshalledFunc{reflect.ValueOf(decisionFunc)}.DecisionFunc
}

func (t *TypedFuncs) MultiDecisionFunc(decisionFunc interface{}) MultiDecisionFunc {
	typeCheck(decisionFunc, []string{"*swf.FSMContext", "swf.HistoryEvent", t.typeArg()}, []string{"[]swf.Decision"})
	return MarshalledFunc{reflect.ValueOf(decisionFunc)}.MultiDecisionFunc
}

func (t *TypedFuncs) StateFunc(stateFunc interface{}) StateFunc {
	typeCheck(stateFunc, []string{"*swf.FSMContext", "swf.HistoryEvent", t.typeArg()}, []string{})
	return MarshalledFunc{reflect.ValueOf(stateFunc)}.StateFunc
}

func (t *TypedFuncs) PredicateFunc(stateFunc interface{}) PredicateFunc {
	typeCheck(stateFunc, []string{t.typeArg()}, []string{"bool"})
	return MarshalledFunc{reflect.ValueOf(stateFunc)}.PredicateFunc
}

type MarshalledFunc struct {
	v reflect.Value
}

func (m MarshalledFunc) Decider(f *FSMContext, h HistoryEvent, data interface{}) Outcome {
	ret := m.v.Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(h), reflect.ValueOf(data)})[0]
	if ret.IsNil() {
		return nil
	}
	return ret.Interface().(Outcome)
}

func (m MarshalledFunc) DecisionFunc(f *FSMContext, h HistoryEvent, data interface{}) Decision {
	ret := m.v.Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(h), reflect.ValueOf(data)})[0]
	return ret.Interface().(Decision)
}

func (m MarshalledFunc) MultiDecisionFunc(f *FSMContext, h HistoryEvent, data interface{}) []Decision {
	ret := m.v.Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(h), reflect.ValueOf(data)})[0]
	return ret.Interface().([]Decision)
}

func (m MarshalledFunc) StateFunc(f *FSMContext, h HistoryEvent, data interface{}) {
	m.v.Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(h), reflect.ValueOf(data)})
}

func (m MarshalledFunc) PredicateFunc(data interface{}) bool {
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
