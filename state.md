# state / structs

each workflow has a main 'data' struct that is threaded through, which is equivalent to a db row in an orm. 'App' etc
workflow versions are then for migrating structs?

each start of a workflow has a StartState{"state":"optional-start-state", "data":"json struct", "input":"json event"}, or is that just on a per FSM basis?

the state field there can be used as to differentiate which FSMState to use on start. start state or continueAsNewState.

each FSM then has the EmptyData func which is for the main 'data' struct, and also now gets a EmptyInputOrResult which is
a func(HistoryEvent) interface{} ?

```
EmptyInputOrResult: func (h HistoryEvent) interface{} {
   switch h.EventType {
   case swf.WokrflowExecutionSignaledEvent:
     return new(SignalStruct)
   }  ///etc...

}
```

then does the Decider become

```
func(f FMS, h HistoryEvent, workflowData interface{}) Outcome {
    switch h.EventType {
       case swf.WokrflowExecutionSignaledEvent:
         signalStruct := f.EventData(h).(*SignalStruct)
       }  ///etc...
}
```

may want to let FSMStates also supply an EmptyDataOrResult that overrides the FSM if present.

also good to pass the fsm through, so states can transition to other states mid event if it makes sense?

make deciders panic safe and go to error state?


# canned FSMStates

"done" ?
"error" state where we go when err happens?