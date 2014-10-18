swf4go
======

go client library for amazon simple workflow service

need a request/response struct?
-------------------------------

* check out json/readme.md
* put a file called json/NameOfActionRequest.json and json/NameOfActionResponse.json
* run jsongen json/NameOfActionRequest.json
* add the struct, after any cleanups/deduping to protocol.go
* add a method to the right client interface and an implementation in the client