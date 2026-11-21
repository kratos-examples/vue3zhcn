package main

import (
	"path/filepath"

	"github.com/yylego/done"
	"github.com/yylego/kratos-ts-http/kratostshttp"
	"github.com/yylego/osexistpath/osmustexist"
	"github.com/yylego/runpath"
	"github.com/yylego/zaplog"
	"go.uber.org/zap"
)

// 从两个后端的 proto 生成前端可直接使用的 TypeScript 客户端。
// Generates the frontend TypeScript clients from the protos of both backends.
//
// 这里走的是 kratos-ts-http 的第二条路线：在 Go 里驱动 buf，落点作为参数传进去，
// 所以跨目录的相对路径由 runpath 算出来，不必写进 Makefile。
// This takes the second kratos-ts-http route: drive buf from Go and pass the destination in,
// so runpath works out the cross-directory path instead of the Makefile carrying it.
//
// 各后端目录里的 `make web_rpc_ts` 是同一件事的第一条路线，产出一致，按手边方便挑一个。
// The `make web_rpc_ts` target in each backend does the same thing through the first route;
// both produce the same output, so take whichever is at hand.
func main() {
	zaplog.SUG.Infoln("=== Vue3 client code gen start ===")

	frontendRoot := osmustexist.ROOT(runpath.PARENT.UpTo(1, "vue3project"))

	runGenerate(runpath.PARENT.UpTo(1, "demo1kratos"), filepath.Join(frontendRoot, "src", "rpc", "demo1"))
	runGenerate(runpath.PARENT.UpTo(1, "demo2kratos"), filepath.Join(frontendRoot, "src", "rpc", "demo2"))

	zaplog.SUG.Infoln("=== finished ===")
}

// runGenerate 生成一个后端的客户端，一次调用完成生成与改写，中间不落任何产物。
// runGenerate covers one backend: a single call does generation and rewriting, with nothing
// staged in between.
//
// `--clean` 让 buf 先清空落点再生成：buf 只覆盖同名文件、从不删除不再生成的旧文件，
// 少了它，proto 一旦不再 import 某个类型，那个 .ts 就会一直躺在前端目录里没人发现。
// ⛔仅因为这个落点【由本次生成独占】才敢清；若某个目录要接纳多趟生成，就不能带它。
// `--clean` wipes the destination first: buf overwrites files but never deletes ones it no
// longer generates, so a type dropped from the proto leaves its .ts behind for good.
// ⛔Safe here only because this destination belongs to this one run.
func runGenerate(kratosRoot string, outPath string) {
	kratosRoot = osmustexist.ROOT(kratosRoot)
	zaplog.LOG.Debug("paths", zap.String("backend", kratosRoot), zap.String("out-path", outPath))

	done.Done(kratostshttp.BufGenerate(kratosRoot, outPath, []string{
		"--template", "buf.gen.ts.yaml",
		"--include-imports",
		"--clean",
	}))

	zaplog.SUG.Infoln("client code sits in", outPath)
}
