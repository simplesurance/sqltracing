# sqltracing

sqltracing is a Go package for tracing database operations via an OpenTracing
tracer.
It can wrap any `driver.Driver` compatible SQL driver.


## History

sqltracing is build on top of the [ngrok/sqlmw](https://github.com/ngrok/sqlmw)
SQL middleware. \
[ngrok/sqlmw](https://github.com/ngrok/sqlmw) is based on
[luna-duclos/instrumentedsql](https://github.com/luna-duclos/instrumentedsql)
which is a fork of
[ExpansiveWorlds(instrumentedsql](https://github.com/ExpansiveWorlds/instrumentedsql).
