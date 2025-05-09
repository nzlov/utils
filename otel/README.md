# Otel

使用`go.opentelemetry.io`追踪数据。可查看例子`ex`

## 使用方法

- 创建`otel`包及`otel.go`文件，注意修改`otelName`

```

package otel

import (
 "go.opentelemetry.io/contrib/bridges/otelslog"
 "go.opentelemetry.io/otel"
)

const otelName = "utils"

var (
 Tracer = otel.Tracer(otelName)
 Start  = Tracer.Start
)

var (
 Log = otelslog.NewLogger(otelName)

 With  = Log.With
 Info  = Log.InfoContext
 Error = Log.ErrorContext
)

var Meter = otel.Meter(otelName)
```

- 创建结构实现`Run`接口

```
type App struct {
 srv *http.Server
}

func NewApp() *App {
 return &App{}
}

func (a *App) Run() error {
 a.srv = &http.Server{
  Addr: ":9999",
  Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
   ctx, span := otel.Tracer.Start(r.Context(), "handler")
   defer span.End()

   otel.Info(ctx, r.URL.String())
  }),
 }

 return a.srv.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
 return a.srv.Shutdown(ctx)
}
```

- 初始化创建`otel.Config`变量

```
 cfg := &otel.Config{}
```

- 使用`otel.Config`变量运行`Run``接口实例

```
 if err := cfg.Run(&App{}); err != nil {
  log.Fatal(err)
 }
```

- 需要追踪使用

```
ctx, span := otel.Tracer.Start(r.Context(), "handler")
defer span.End()
```

- 日志

```
otel.Info(ctx, r.URL.String())
otel.Error(ctx, r.URL.String())
```

- 多条日志附加参数

```
log := otel.With("node", "123")
log.InfoContext(ctx, " Start")
```

- 启动项目使用

- 如果需要发送到`http`,例如`openobserve`使用`go.opentelemetry.io`环境变量

```
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:5080/api/default OTEL_RESOURCE_ATTRIBUTES="service.name=xxx,service.version=0.1.0" OTEL_EXPORTER_OTLP_HEADERS="stream-name=xxx,Authorization=Basic xxx"
```
