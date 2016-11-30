[![GoDoc](https://godoc.org/github.com/lovoo/gcloud-opentracing?status.svg)](http://godoc.org/github.com/lovoo/gcloud-opentracing)
# gcloud-opentracing
 OpenTracing Tracer implementation for GCloud StackDriver in Go. Based on [basictracer](https://github.com/opentracing/basictracer-go) and implemented `Recorder` for this propose.
 
### Getting Started
-------------------
To install fcm, use `go get`:

```bash
go get github.com/lovoo/gcloud-opentracing
```
or `govendor`:

```bash
govendor fetch github.com/lovoo/gcloud-opentracing
```
or other tool for vendoring.

### Sample Usage
-------------------
First of all, you need to init Global Tracer with GCloud Tracer:
```go
package main

import (
    gcloudtracer "github.com/lovoo/gcloud-opentracing"
    opentracing "github.com/opentracing/opentracing-go"
    "golang.org/x/net/context"
)

func main() {
    // ...
    opentracing.InitGlobalTracer(context.Background(), gcloudtracer.WithProject("project-id"))
    // ...
}
```

Then you can create traces as decribed [here](https://github.com/opentracing/opentracing-go). More information you can find on [OpenTracing project](http://opentracing.io) website.
