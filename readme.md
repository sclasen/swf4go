swf4go
======

go client library for amazon simple workflow service

[![Build Status](https://travis-ci.org/sclasen/swf4go.svg?branch=master)](https://travis-ci.org/sclasen/swf4go)

* godoc here: http://godoc.org/github.com/sclasen/swf4go

swf4go provides a full client library for amazon simple workflow service, as well as Pollers for both ActivityTasks and DecisionTasks which work around
some of the eccentricities of the swf service.

swf4go also uses these facilities to provide an erlang/akka style Finite State Machine abstraction, which is used to model workflows as FSMs.

Please see the godoc for detailed documentation and examples.
