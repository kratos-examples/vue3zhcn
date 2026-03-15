package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/yylego/done"
	"github.com/yylego/kratos-vue3/vue3kratos"
	"github.com/yylego/must"
	"github.com/yylego/osexec"
	"github.com/yylego/osexistpath/osmustexist"
	"github.com/yylego/rese"
	"github.com/yylego/runpath"
	"github.com/yylego/zaplog"
	"go.uber.org/zap"
)

func main() {
	zaplog.SUG.Infoln("=== Vue3 Client Code Gen Workflow Start ===")

	frontendRoot := runpath.PARENT.UpTo(1, "vue3project")

	// Generate for demo1kratos
	demo1Root := runpath.PARENT.UpTo(1, "demo1kratos")
	zaplog.LOG.Debug("demo1kratos", zap.String("backend", demo1Root), zap.String("frontend", frontendRoot))
	runGenerate(demo1Root, filepath.Join(frontendRoot, "src/rpc/demo1"))

	// Generate for demo2kratos
	demo2Root := runpath.PARENT.UpTo(1, "demo2kratos")
	zaplog.LOG.Debug("demo2kratos", zap.String("backend", demo2Root), zap.String("frontend", frontendRoot))
	runGenerate(demo2Root, filepath.Join(frontendRoot, "src/rpc/demo2"))

	zaplog.SUG.Infoln("=== WORKFLOW FINISHED SUCCESS! ===")
}

func runGenerate(kratosRoot string, clientCodeDest string) {
	// Step 1: Validate backend project
	zaplog.SUG.Infoln("Backend project:", kratosRoot)
	osmustexist.ROOT(kratosRoot)

	// Step 2: Check Makefile exists and contains required targets
	makefilePath := filepath.Join(kratosRoot, "Makefile")
	osmustexist.FILE(makefilePath)

	makefileData := rese.A1(os.ReadFile(makefilePath))
	must.True(bytes.Contains(makefileData, []byte("web_api_grpc_ts:")))
	must.True(bytes.Contains(makefileData, []byte("web_api_grpc_to_http:")))
	must.True(bytes.Contains(makefileData, []byte("web_api_cleanup:")))
	zaplog.SUG.Infoln("Makefile targets verified")

	// Step 3: Generate TypeScript gRPC clients from proto files
	grpcTsOutput := filepath.Join(kratosRoot, "bin", "web_api_grpc_ts.out")
	zaplog.SUG.Infoln("Generating TypeScript gRPC clients...")
	zaplog.SUG.Infoln("   Output DIR:", grpcTsOutput)

	if osmustexist.IsRootExist(filepath.Join(kratosRoot, "bin")) {
		zaplog.SUG.Infoln("   Cleaning previous output...")
		done.Done(os.RemoveAll(grpcTsOutput))
	}

	rese.A1(osexec.ExecInPath(kratosRoot, "make", "web_api_grpc_ts"))
	osmustexist.ROOT(grpcTsOutput)
	zaplog.SUG.Infoln("TypeScript gRPC clients generated")

	// Step 4: Convert gRPC clients to HTTP clients (in bin directory)
	zaplog.SUG.Infoln("Converting gRPC clients to HTTP clients...")
	rese.A1(osexec.ExecInPath(kratosRoot, "make", "web_api_grpc_to_http"))
	zaplog.SUG.Infoln("Conversion completed")

	// Step 5: Sync converted files to frontend project
	osmustexist.ROOT(clientCodeDest)

	zaplog.SUG.Infoln("Syncing converted files...")
	zaplog.SUG.Infoln("   From:", grpcTsOutput)
	zaplog.SUG.Infoln("   To:  ", clientCodeDest)
	vue3kratos.CloneFilesToDestRoot(grpcTsOutput, clientCodeDest)
	zaplog.SUG.Infoln("File sync completed")

	// Step 6: Cleanup temp files
	zaplog.SUG.Infoln("Cleaning up temp files...")
	rese.A1(osexec.ExecInPath(kratosRoot, "make", "web_api_cleanup"))
	zaplog.SUG.Infoln("Cleanup completed")
}
