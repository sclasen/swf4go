/*
Package swf provides a full implementation of a client api for Amazon Simple Workflow Service
http://docs.aws.amazon.com/amazonswf/latest/apireference/

In addition it provides a basic facility for modeling swf workflows as FSMs (finite state machines), as well as
implementations of pollers for both decision and activty tasks.

Client

The Client in this library understands how to make requests to and receive responses from every endpoint in the SWF API, as well as
the PutRecord endpoint for Kinesis. (This will be explained later).

TaskPollers

DecisionTaskPoller and ActivityTaskPoller facilitate proper usage of the PollForDecisionTask and PollForActivityTask endpoints in the SWF API. These endpoints are used by
DecisionTask and ActivityTask workers to claim tasks on which to work. The endpoints use long polling. SWF will keep the request open for up to
60 seconds before returning an 'empty' response. If a task is generated before that time, a non-empty task is delivered (and assigned to) a particular
polling client.

There is an unfortunate bug in SWF that occurs when a long-polling request gets terminated client side, rather than waiting for the SWF API to respond.
SWF does not recognize this condition so it can result in assigning a task to a disconnected worker, which will subsequently cause the task to timeout.
This is not terrible if the task has a short timeout but can cause big delays if the task does have a long timeout.

Both types of pollers allow you to manage polling yourself by calling Poll() directly. However it is recommended that you use the
PollUntilShutdownBy(...) function, which works in concert with a PollerShutdownManager to handle os.Signals properly and await all in-flight polls to complete before allowing
a process to exit.

PollerShutdownManager

The PollerShutdownManager, once created by calling swf.RegisterPollerShutdownManager() will handle SIGHUP,SIGINT,SIGTERM,SIGQUIT by signalling any registered pollers to shut down
once any in-flight polls have completed. Once this process is complete, the PollerShutdownManager will call os.Exit(0), causing the process to terminate. The shutdown process can take up to 60 seconds
due to the length of SWF long polls before an empty response is returned.

FSM

The FSM in swf4go layers an erlang/akka style finite state machine abstraction on top of SWF, and facilitates modelling your workflows as FSMs. The FSM will be responsible for handling the decsison
tasks in your workflow that implicitly model it.

The FSM takes care of serializing/deserializing and  threading a data model through the workflow history for you, as well as serialization/deserialization of any payloads in events your workflows recieve,
as well as optionally sending the data model snapshots to kinesis, to facilitate a CQRS style application where the query models will be built off the Kinesis stream.

From http://www.erlang.org/doc/design_principles/fsm.html, A finite state machine, FSM, can be described as a set of relations of the form:

    State(S) x Event(E) -> Actions(A), State(S')

Substituting the relevant swf/swf4go concepts, we get

   (Your main data struct) x (an swf.HistoryEvent) -> (zero or more swf.Decisions), (A possibly updated main data struct)


*/
package swf





