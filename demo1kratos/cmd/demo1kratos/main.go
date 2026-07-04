package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/go-kratos/kratos/contrib/encoding/json/v3"
	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/config"
	"github.com/go-kratos/kratos/v3/config/file"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"
	"github.com/yylego/done"
	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
	"github.com/yylego/must"
	"github.com/yylego/rese"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")

	// Configure JSON field naming style for HTTP responses
	// UseProtoNames=false uses lowerCamelCase to work across different languages
	// 配置 HTTP 响应的 JSON 字段命名风格，使用小写驼峰命名确保跨语言兼容性
	json.MarshalOptions.UseProtoNames = false

	// Set UseEnumNumbers to true to serialize enums as numbers instead of strings
	// This matches TypeScript generated code from proto, ensuring frontend works correct
	// 设置 UseEnumNumbers 为 true 使枚举序列化为数字而非字符串，与 proto 生成的 TypeScript 代码保持一致
	json.MarshalOptions.UseEnumNumbers = true
}

func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(done.VCE(os.Hostname()).Omit()),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()
	logger := log.NewLogger(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		}),
		log.WithExtractor(tracing.TraceAttrs),
	).With(
		slog.String("service.id", done.VCE(os.Hostname()).Omit()),
		slog.String("service.name", Name),
		slog.String("service.version", Version),
	)
	log.SetDefault(logger)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer rese.F0(c.Close)

	must.Done(c.Load())

	var cfg conf.Bootstrap
	must.Done(c.Scan(&cfg))

	app, cleanup := rese.V2(wireApp(cfg.Server, cfg.Data, logger))
	defer cleanup()

	// start and wait for stop signal
	must.Done(app.Run())
}
