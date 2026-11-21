# Changes

Code differences compared to source project.

## Makefile (+22 -0)

```diff
@@ -7,6 +7,8 @@
 init:
 	go install github.com/google/wire/cmd/wire@latest
 	go install github.com/bufbuild/buf/cmd/buf@latest
+	go install github.com/yylego/kratos-ts-http/cmd/protoc-gen-kratos-ts-http@latest
+	npm install -g @protobuf-ts/plugin
 
 .PHONY: config
 # generate internal proto
@@ -35,6 +37,26 @@
 	make api
 	make config
 	make generate
+
+# 从 proto 一步生成浏览器可直接使用的 TypeScript 客户端，产出即是最终代码，无需二次改写。
+# 落到本项目的 bin 下，之后人工拷进前端；`.out` 结尾的目录 Go 的 .gitignore 本就忽略。
+# generate web-safe TypeScript clients from proto; the output is the finished code.
+# It lands under bin here, ready to be copied into the frontend by hand.
+#
+# 依赖的两个插件由 `make init` 一并安装。
+# Both plugins it needs come from `make init`.
+#
+# 想一步直达前端，跑 vue3codegen 那条（路径由 Go 算，不必写死在这里）。
+# For a straight run into the frontend, use vue3codegen: it works the path out in Go.
+web_rpc_ts:
+	buf generate --template buf.gen.ts.yaml --include-imports --clean -o ./bin/web_rpc_ts.out
+
+# 删掉本地产物目录本身。平时用不上——生成那条带 `--clean`，每次已是从零开始，
+# 而 `.out` 结尾的目录 git 本就忽略；只在想彻底腾空这块磁盘时才用它。
+# Removes the local output directory itself. Rarely needed: the generate target passes `--clean`,
+# so each run already starts from nothing, and git ignores `.out` dirs anyway.
+web_rpc_ts_cleanup:
+	rm -rf ./bin/web_rpc_ts.out
 
 # show help
 help:
```

## buf.gen.ts.yaml (+14 -0)

```diff
@@ -0,0 +1,14 @@
+# 使用 kratos-ts-http 的插件，产出即为最终的 HTTP 形态客户端，无需第二步改写。
+# 落点不在此处指定，由命令行的 -o 决定。
+# Uses the kratos-ts-http plugin, so the output is the finished HTTP-form client.
+# The destination is set on the command line with -o, not here.
+version: v2
+inputs:
+  - directory: api
+plugins:
+  - local: protoc-gen-kratos-ts-http
+    out: .
+    opt:
+      - ts_nocheck
+      - eslint_disable
+      - long_type_string
```

## cmd/demo2kratos/main.go (+12 -0)

```diff
@@ -5,6 +5,7 @@
 	"log/slog"
 	"os"
 
+	"github.com/go-kratos/kratos/contrib/encoding/json/v3"
 	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
 	"github.com/go-kratos/kratos/v3"
 	"github.com/go-kratos/kratos/v3/config"
@@ -31,6 +32,17 @@
 
 func init() {
 	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
+
+	// Configure JSON field naming style for HTTP responses
+	// UseProtoNames=false uses lowerCamelCase to work across different languages
+	// 配置 HTTP 响应的 JSON 字段命名风格，使用小写驼峰命名确保跨语言兼容性
+	json.MarshalOptions.UseProtoNames = false
+
+	// Set UseEnumNumbers to true to serialize enums as numbers instead of strings
+	// This matches TypeScript generated code from proto, ensuring frontend works correct
+	// 设置 UseEnumNumbers 为 true 使枚举序列化为数字而非字符串，与 proto 生成的 TypeScript 代码保持一致
+	json.MarshalOptions.UseEnumNumbers = true
+
 }
 
 func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
```

## cmd/demo2kratos/wire_gen.go (+4 -4)

```diff
@@ -24,14 +24,14 @@
 	if err != nil {
 		return nil, nil, err
 	}
-	articleUsecase, err := biz.NewArticleUsecase(dataData, logger)
+	uc文章管理, err := biz.NewUc文章管理(dataData, logger)
 	if err != nil {
 		cleanup()
 		return nil, nil, err
 	}
-	articleService := service.NewArticleService(articleUsecase)
-	grpcServer := server.NewGRPCServer(confServer, articleService, logger)
-	httpServer := server.NewHTTPServer(confServer, articleService, logger)
+	svc文章服务 := service.NewSvc文章服务(uc文章管理)
+	grpcServer := server.NewGRPCServer(confServer, svc文章服务, logger)
+	httpServer := server.NewHTTPServer(confServer, svc文章服务, logger)
 	app := newApp(logger, grpcServer, httpServer)
 	return app, func() {
 		cleanup()
```

## internal/biz/article.go (+202 -122)

```diff
@@ -2,187 +2,267 @@
 
 import (
 	"context"
-	"errors"
 	"log/slog"
 
+	"github.com/go-kratos/kratos/v3/errors"
+	"github.com/yylego/gormcnm"
+	"github.com/yylego/gormrepo"
+	"github.com/yylego/gormrepo/gormclass"
 	"github.com/yylego/kratos-ebz/ebzkratos"
 	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
+	"github.com/yylego/kratos-examples/demo2kratos/internal/pkg/models"
+	"github.com/yylego/kratos-gorm/gormkratos"
 	"github.com/yylego/must"
 	"gorm.io/gorm"
 	"gorm.io/gorm/clause"
 )
 
-// Article is the GORM type mapped to the "articles" table. This service owns
-// the table; demo1kratos keeps a duplicate of it just to cascade-delete a
-// student's articles (the two services share one database).
-//
-// Article 是映射到 articles 表的 GORM 模型，本服务是这张表的归属方；
-// demo1kratos 里有一份镜像，仅用于删学生时顺带删文章（两服务共用一个库）
-type Article struct {
-	ID        int64  `gorm:"primaryKey;autoIncrement"`
-	Title     string `gorm:"size:256;not null"`
-	Content   string `gorm:"type:text"`
-	StudentID int64  `gorm:"index"`
+type Req文章信息 struct {
+	ID    int64
+	V标题   string
+	V内容   string
+	V学生编号 int64
 }
 
-func (Article) TableName() string { return "articles" }
-
-type ArticleUsecase struct {
-	data *data.Data
-	slog *slog.Logger
+type Uc文章管理 struct {
+	data   *data.Data
+	repo文章 *gormrepo.Repo[models.T文章, *models.T文章Columns]
+	repo学生 *gormrepo.Repo[models.T学生, *models.T学生Columns]
+	slog   *slog.Logger
 }
 
-func NewArticleUsecase(data *data.Data, logger *slog.Logger) (*ArticleUsecase, error) {
-	// Migrate the owned table plus the mirrored students table (needed in the
-	// existence check); both services share one database
-	// 建好本服务拥有的 articles 表，外加镜像的 students 表（供存在性校验用）
-	if err := data.DB().AutoMigrate(&Article{}, &Student{}); err != nil {
+func NewUc文章管理(data *data.Data, logger *slog.Logger) (*Uc文章管理, error) {
+	// 与学生服务共用一个库：在这里把两张表都建好。
+	if err := data.DB().AutoMigrate(&models.T文章{}, &models.T学生{}); err != nil {
 		return nil, err
 	}
-	return &ArticleUsecase{data: data, slog: logger}, nil
+	return &Uc文章管理{
+		data:   data,
+		repo文章: gormrepo.NewRepo(gormclass.Use(&models.T文章{})),
+		repo学生: gormrepo.NewRepo(gormclass.Use(&models.T学生{})),
+		slog:   logger,
+	}, nil
 }
 
-func (uc *ArticleUsecase) CreateArticle(ctx context.Context, a *Article) (*Article, *ebzkratos.Ebz) {
-	must.Nice(a.Title)
-	must.True(a.StudentID > 0)
+func (uc *Uc文章管理) Xqt创建文章(ctx context.Context, req *Req文章信息) (*Req文章信息, *ebzkratos.Ebz) {
+	must.Nice(req.V标题)
+	must.True(req.V学生编号 > 0)
 
-	// Lock the student row and insert the article in one transaction: the FOR
-	// SHARE lock blocks a concurrent DeleteStudent (which takes FOR UPDATE) from
-	// removing this student before we commit, so we cannot end up with an article
-	// pointing at a student that's being deleted.
-	// 在一个事务里锁住学生行再插入文章：FOR SHARE 锁会挡住并发的 DeleteStudent
-	// （它持 FOR UPDATE）在本事务提交前删除该学生，从而绝不会创建出指向
-	// "正在被删除的学生"的文章
-	res := &Article{Title: a.Title, Content: a.Content, StudentID: a.StudentID}
-	err := uc.data.DB().WithContext(ctx).Transaction(func(db *gorm.DB) error {
-		var student Student
-		if err := db.Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).First(&student, a.StudentID).Error; err != nil {
-			return err
+	db := uc.data.DB()
+
+	var v文章 *models.T文章
+
+	// 翻译桩子：在一个事务里 FOR SHARE 锁住学生行再插文章，挡住并发的
+	// DeleteStudent（它持 FOR UPDATE），绝不创建指向"正被删除的学生"的文章。
+	if erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
+		if _, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
+			return db.Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).Where(cls.ID.Eq(req.V学生编号))
+		}); erb != nil {
+			if erb.NotExist {
+				return pb.ErrorBadParam("student %d does not exist", req.V学生编号)
+			}
+			return pb.ErrorDbError("get student: %v", erb.Cause)
 		}
-		return db.Create(res).Error
-	})
-	if err != nil {
-		if errors.Is(err, gorm.ErrRecordNotFound) {
-			return nil, ebzkratos.New(pb.ErrorBadParam("student %d does not exist", a.StudentID))
+		v文章 = &models.T文章{
+			V标题:   req.V标题,
+			V内容:   req.V内容,
+			V学生编号: req.V学生编号,
 		}
-		return nil, ebzkratos.New(pb.ErrorArticleCreateFailure("create article: %v", err))
+		if err := uc.repo文章.With(ctx, db).Create(v文章); err != nil {
+			return pb.ErrorArticleCreateFailure("create article: %v", err)
+		}
+		return nil
+	}); err != nil {
+		if erk != nil {
+			return nil, ebzkratos.New(erk)
+		}
+		return nil, ebzkratos.New(pb.ErrorTxError("tx: %v", err))
 	}
-	uc.slog.InfoContext(ctx, "created article", "id", res.ID, "student_id", res.StudentID)
+	res := &Req文章信息{
+		ID:    int64(v文章.ID),
+		V标题:   v文章.V标题,
+		V内容:   v文章.V内容,
+		V学生编号: v文章.V学生编号,
+	}
+	uc.slog.InfoContext(ctx, "created article", "id", res.ID, "student_id", res.V学生编号)
 	return res, nil
 }
 
-func (uc *ArticleUsecase) UpdateArticle(ctx context.Context, a *Article) (*Article, *ebzkratos.Ebz) {
-	must.True(a.ID > 0)
-	must.Nice(a.Title)
-	must.True(a.StudentID > 0)
+func (uc *Uc文章管理) Xqt更新文章(ctx context.Context, req *Req文章信息) (*Req文章信息, *ebzkratos.Ebz) {
+	must.True(req.ID > 0)
+	must.Nice(req.V标题)
+	must.True(req.V学生编号 > 0)
 
-	// Same transaction + FOR SHARE lock as CreateArticle: the (new) owning
-	// student cannot be deleted while we re-point the article.
-	// 与 CreateArticle 相同的事务 + FOR SHARE 锁：改文章归属期间，新归属的学生不会被并发删除
-	res := &Article{ID: a.ID}
-	var studentMissing, articleMissing bool
-	err := uc.data.DB().WithContext(ctx).Transaction(func(db *gorm.DB) error {
-		var student Student
-		if err := db.Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).First(&student, a.StudentID).Error; err != nil {
-			if errors.Is(err, gorm.ErrRecordNotFound) {
-				studentMissing = true
-				return nil
+	db := uc.data.DB()
+
+	var v文章 *models.T文章
+	// FOR SHARE 锁住新归属学生（对齐桩子），再更新文章、靠 RowsAffected 判文章不存在，最后 reload——忠实对齐桩子。
+	if erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
+		if _, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
+			return db.Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).Where(cls.ID.Eq(req.V学生编号))
+		}); erb != nil {
+			if erb.NotExist {
+				return pb.ErrorBadParam("student %d does not exist", req.V学生编号)
 			}
-			return err
+			return pb.ErrorDbError("get student: %v", erb.Cause)
 		}
-		upd := db.Model(res).Updates(map[string]any{
-			"title":      a.Title,
-			"content":    a.Content,
-			"student_id": a.StudentID,
+		upd := uc.repo文章.Wrap(ctx, db).UpdatesM(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+			return db.Where(cls.ID.Eq(req.ID))
+		}, func(cls *models.T文章Columns) gormcnm.ColumnValueMap {
+			return cls.Kw(cls.V标题.Kv(req.V标题)).Kw(cls.V内容.Kv(req.V内容)).Kw(cls.V学生编号.Kv(req.V学生编号))
 		})
 		if upd.Error != nil {
-			return upd.Error
+			return pb.ErrorDbError("update article: %v", upd.Error)
 		}
 		if upd.RowsAffected == 0 {
-			articleMissing = true
-			return nil
+			return pb.ErrorArticleNotFound("article %d not found", req.ID)
 		}
-		return db.First(res, a.ID).Error
-	})
-	if err != nil {
-		return nil, ebzkratos.New(pb.ErrorDbError("update article: %v", err))
+		got, erb := uc.repo文章.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+			return db.Where(cls.ID.Eq(req.ID))
+		})
+		if erb != nil {
+			return pb.ErrorDbError("reload article: %v", erb.Cause)
+		}
+		v文章 = got
+		return nil
+	}); err != nil {
+		if erk != nil {
+			return nil, ebzkratos.New(erk)
+		}
+		return nil, ebzkratos.New(pb.ErrorTxError("tx: %v", err))
 	}
-	if studentMissing {
-		return nil, ebzkratos.New(pb.ErrorBadParam("student %d does not exist", a.StudentID))
-	}
-	if articleMissing {
-		return nil, ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", a.ID))
-	}
-	return res, nil
+
+	return &Req文章信息{
+		ID:    int64(v文章.ID),
+		V标题:   v文章.V标题,
+		V内容:   v文章.V内容,
+		V学生编号: v文章.V学生编号,
+	}, nil
 }
 
-func (uc *ArticleUsecase) DeleteArticle(ctx context.Context, id int64) *ebzkratos.Ebz {
-	must.True(id > 0)
+func (uc *Uc文章管理) Xqt删除文章(ctx context.Context, n编号 int64) *ebzkratos.Ebz {
+	must.True(n编号 > 0)
 
-	del := uc.data.DB().WithContext(ctx).Delete(&Article{}, id)
+	db := uc.data.DB()
+
+	// 忠实翻译桩子：单条删除后靠 RowsAffected 判不存在（0 行→ArticleNotFound）。
+	del := uc.repo文章.Wrap(ctx, db).DeleteW(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+		return db.Where(cls.ID.Eq(n编号))
+	})
 	if del.Error != nil {
 		return ebzkratos.New(pb.ErrorDbError("delete article: %v", del.Error))
 	}
 	if del.RowsAffected == 0 {
-		return ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", id))
+		return ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", n编号))
 	}
-	uc.slog.InfoContext(ctx, "deleted article", "id", id)
+	uc.slog.InfoContext(ctx, "deleted article", "id", n编号)
 	return nil
 }
 
-func (uc *ArticleUsecase) GetArticle(ctx context.Context, id int64) (*Article, *ebzkratos.Ebz) {
-	must.True(id > 0)
+func (uc *Uc文章管理) Get获取文章(ctx context.Context, n编号 int64) (*Req文章信息, *ebzkratos.Ebz) {
+	must.True(n编号 > 0)
 
-	res := &Article{}
-	if err := uc.data.DB().WithContext(ctx).First(res, id).Error; err != nil {
-		if errors.Is(err, gorm.ErrRecordNotFound) {
-			return nil, ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", id))
+	db := uc.data.DB()
+
+	v文章, erb := uc.repo文章.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+		return db.Where(cls.ID.Eq(n编号))
+	})
+	if erb != nil {
+		if erb.NotExist {
+			return nil, ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", n编号))
 		}
-		return nil, ebzkratos.New(pb.ErrorDbError("get article: %v", err))
+		return nil, ebzkratos.New(pb.ErrorDbError("get article: %v", erb.Cause))
 	}
-	return res, nil
+
+	return &Req文章信息{
+		ID:    int64(v文章.ID),
+		V标题:   v文章.V标题,
+		V内容:   v文章.V内容,
+		V学生编号: v文章.V学生编号,
+	}, nil
 }
 
-func (uc *ArticleUsecase) ListArticles(ctx context.Context, page int32, pageSize int32) ([]*Article, int32, *ebzkratos.Ebz) {
-	must.True(page >= 1)
-	must.True(pageSize >= 1)
+// Req文章列表 / Res文章列表 分别包裹「查询文章列表」的入参（分页）与返回（当页数据+总数）。
+type Req文章列表 struct {
+	V页码   int32
+	V每页大小 int32
+}
 
-	db := uc.data.DB().WithContext(ctx)
+type Res文章列表 struct {
+	V文章列表 []*Req文章信息
+	V总数   int32
+}
 
-	var count int64
-	if err := db.Model(&Article{}).Count(&count).Error; err != nil {
-		return nil, 0, ebzkratos.New(pb.ErrorDbError("count articles: %v", err))
-	}
+func (uc *Uc文章管理) Get文章列表(ctx context.Context, req *Req文章列表) (*Res文章列表, *ebzkratos.Ebz) {
+	must.True(req.V页码 >= 1)
+	must.True(req.V每页大小 >= 1)
 
-	var items []*Article
-	if err := db.Order("id").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&items).Error; err != nil {
-		return nil, 0, ebzkratos.New(pb.ErrorDbError("list articles: %v", err))
+	db := uc.data.DB()
+
+	// gormrepo FindPageAndCount 一次拿到当页数据和总行数，对齐桩子的分页+计数。
+	v文章们, total, err := uc.repo文章.With(ctx, db).FindPageAndCount(
+		func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+			return db
+		},
+		func(cls *models.T文章Columns) gormcnm.OrderByBottle {
+			return cls.ID.Ob("asc")
+		},
+		&gormrepo.Pagination{
+			Offset: int((req.V页码 - 1) * req.V每页大小),
+			Limit:  int(req.V每页大小),
+		},
+	)
+	if err != nil {
+		return nil, ebzkratos.New(pb.ErrorDbError("list articles: %v", err))
 	}
-	return items, int32(count), nil
+
+	return &Res文章列表{V文章列表: conv文章列表(v文章们), V总数: int32(total)}, nil
 }
 
-// ListStudentArticles returns one student's articles, one page at a time. The
-// student↔article relationship gets its own endpoint instead of overloading
-// ListArticles with an extra flag.
-//
-// ListStudentArticles 分页返回某个学生的文章。学生↔文章这层关系单独开一个接口，
-// 而不是往 ListArticles 上塞过滤参数。
-func (uc *ArticleUsecase) ListStudentArticles(ctx context.Context, studentID int64, page int32, pageSize int32) ([]*Article, int32, *ebzkratos.Ebz) {
-	must.True(studentID > 0)
-	must.True(page >= 1)
-	must.True(pageSize >= 1)
+// Req学生文章列表 包裹「查询某学生的文章列表」的入参（学生编号+分页），返回复用 Res文章列表。
+type Req学生文章列表 struct {
+	V学生编号 int64
+	V页码   int32
+	V每页大小 int32
+}
 
-	db := uc.data.DB().WithContext(ctx)
+// Get学生文章列表 分页返回某个学生的文章（对应 proto 的 ListStudentArticles）。
+func (uc *Uc文章管理) Get学生文章列表(ctx context.Context, req *Req学生文章列表) (*Res文章列表, *ebzkratos.Ebz) {
+	must.True(req.V学生编号 > 0)
+	must.True(req.V页码 >= 1)
+	must.True(req.V每页大小 >= 1)
 
-	var count int64
-	if err := db.Model(&Article{}).Where("student_id = ?", studentID).Count(&count).Error; err != nil {
-		return nil, 0, ebzkratos.New(pb.ErrorDbError("count student articles: %v", err))
+	db := uc.data.DB()
+
+	v文章们, total, err := uc.repo文章.With(ctx, db).FindPageAndCount(
+		func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+			return db.Where(cls.V学生编号.Eq(req.V学生编号))
+		},
+		func(cls *models.T文章Columns) gormcnm.OrderByBottle {
+			return cls.ID.Ob("asc")
+		},
+		&gormrepo.Pagination{
+			Offset: int((req.V页码 - 1) * req.V每页大小),
+			Limit:  int(req.V每页大小),
+		},
+	)
+	if err != nil {
+		return nil, ebzkratos.New(pb.ErrorDbError("list student articles: %v", err))
 	}
 
-	var items []*Article
-	if err := db.Where("student_id = ?", studentID).Order("id").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&items).Error; err != nil {
-		return nil, 0, ebzkratos.New(pb.ErrorDbError("list student articles: %v", err))
+	return &Res文章列表{V文章列表: conv文章列表(v文章们), V总数: int32(total)}, nil
+}
+
+func conv文章列表(v文章们 []*models.T文章) []*Req文章信息 {
+	a文章列表 := make([]*Req文章信息, 0, len(v文章们))
+	for _, v := range v文章们 {
+		a文章列表 = append(a文章列表, &Req文章信息{
+			ID:    int64(v.ID),
+			V标题:   v.V标题,
+			V内容:   v.V内容,
+			V学生编号: v.V学生编号,
+		})
 	}
-	return items, int32(count), nil
+	return a文章列表
 }
```

## internal/biz/biz.go (+1 -1)

```diff
@@ -2,4 +2,4 @@
 
 import "github.com/google/wire"
 
-var ProviderSet = wire.NewSet(NewArticleUsecase)
+var ProviderSet = wire.NewSet(NewUc文章管理)
```

## internal/pkg/models/article.go (+12 -0)

```diff
@@ -0,0 +1,12 @@
+package models
+
+type T文章 struct {
+	ID    int64  `gorm:"primaryKey;autoIncrement"`
+	V标题   string `gorm:"column:title;size:256;not null"`
+	V内容   string `gorm:"column:content;type:text"`
+	V学生编号 int64  `gorm:"column:student_id;index:idx_articles_student_id"`
+}
+
+func (*T文章) TableName() string {
+	return "articles"
+}
```

## internal/pkg/models/gormcnm.gen.go (+56 -0)

```diff
@@ -0,0 +1,56 @@
+// Code generated using gormcngen. DO NOT EDIT.
+// This file was auto generated via github.com/yylego/gormcngen
+
+//go:build !gormcngen_generate
+
+// Generated from: gormcnm.gen_test.go:35 -> models_test.TestGenerateColumns
+// ========== GORMCNGEN:DO-NOT-EDIT-MARKER:END ==========
+
+// Code generated using gormcngen. DO NOT EDIT.
+// This file was auto generated via github.com/yylego/gormcngen
+
+package models
+
+import (
+	"github.com/yylego/gormcnm"
+)
+
+func (c *T文章) Columns() *T文章Columns {
+	return &T文章Columns{
+		// Auto-generated: column names and types mapping. DO NOT EDIT. // 自动生成：列名和类型映射。请勿编辑。
+		ID:    gormcnm.Cnm(c.ID, "id"),
+		V标题:   gormcnm.Cnm(c.V标题, "title"),
+		V内容:   gormcnm.Cnm(c.V内容, "content"),
+		V学生编号: gormcnm.Cnm(c.V学生编号, "student_id"),
+	}
+}
+
+type T文章Columns struct {
+	// Auto-generated: embedding operation functions to make it simple to use. DO NOT EDIT. // 自动生成：嵌入操作函数便于使用。请勿编辑。
+	gormcnm.ColumnOperationClass
+	// Auto-generated: column names and types in database table. DO NOT EDIT. // 自动生成：数据库表的列名和类型。请勿编辑。
+	ID    gormcnm.ColumnName[int64]
+	V标题   gormcnm.ColumnName[string]
+	V内容   gormcnm.ColumnName[string]
+	V学生编号 gormcnm.ColumnName[int64]
+}
+
+func (c *T学生) Columns() *T学生Columns {
+	return &T学生Columns{
+		// Auto-generated: column names and types mapping. DO NOT EDIT. // 自动生成：列名和类型映射。请勿编辑。
+		ID:  gormcnm.Cnm(c.ID, "id"),
+		V名字: gormcnm.Cnm(c.V名字, "name"),
+		V年龄: gormcnm.Cnm(c.V年龄, "age"),
+		V班级: gormcnm.Cnm(c.V班级, "class_name"),
+	}
+}
+
+type T学生Columns struct {
+	// Auto-generated: embedding operation functions to make it simple to use. DO NOT EDIT. // 自动生成：嵌入操作函数便于使用。请勿编辑。
+	gormcnm.ColumnOperationClass
+	// Auto-generated: column names and types in database table. DO NOT EDIT. // 自动生成：数据库表的列名和类型。请勿编辑。
+	ID  gormcnm.ColumnName[int64]
+	V名字 gormcnm.ColumnName[string]
+	V年龄 gormcnm.ColumnName[int32]
+	V班级 gormcnm.ColumnName[string]
+}
```

## internal/pkg/models/gormcnm.gen_test.go (+37 -0)

```diff
@@ -0,0 +1,37 @@
+package models_test
+
+import (
+	"testing"
+
+	"github.com/yylego/gormcngen"
+	"github.com/yylego/kratos-examples/demo2kratos/internal/pkg/models"
+	"github.com/yylego/osexistpath/osmustexist"
+	"github.com/yylego/runpath/runtestpath"
+)
+
+// Auto generate columns with go generate command
+// Support execution via: go generate ./...
+// Delete this comment block if auto generation is not needed
+//
+//go:generate go test -v -run TestGenerateColumns
+func TestGenerateColumns(t *testing.T) {
+	// Retrieve the absolute path of the source file based on current test file location
+	absPath := osmustexist.FILE(runtestpath.SrcPath(t))
+	t.Log(absPath)
+
+	// Define data objects used in column generation - supports both instance and non-instance types
+	objects := []any{
+		&models.T文章{},
+		&models.T学生{},
+	}
+
+	// Configure generation options with latest best practices
+	options := gormcngen.NewOptions().
+		WithColumnClassExportable(true). // Generate exportable column class names like T文章Columns
+		WithColumnsMethodRecvName("c").  // Set receiver name for column methods
+		WithColumnsCheckFieldType(true)  // Enable field type checking for type safe
+
+	// Create configuration and generate code to target file
+	cfg := gormcngen.NewConfigs(objects, options, absPath)
+	cfg.Gen() // Generate code to "gormcnm.gen.go" file
+}
```

## internal/pkg/models/student.go (+12 -0)

```diff
@@ -0,0 +1,12 @@
+package models
+
+type T学生 struct {
+	ID  int64  `gorm:"primaryKey;autoIncrement"`
+	V名字 string `gorm:"column:name;size:128;not null"`
+	V年龄 int32  `gorm:"column:age"`
+	V班级 string `gorm:"column:class_name;size:128"`
+}
+
+func (*T学生) TableName() string {
+	return "students"
+}
```

## internal/server/grpc.go (+1 -1)

```diff
@@ -11,7 +11,7 @@
 	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
 )
 
-func NewGRPCServer(c *conf.Server, article *service.ArticleService, logger *slog.Logger) *grpc.Server {
+func NewGRPCServer(c *conf.Server, article *service.Svc文章服务, logger *slog.Logger) *grpc.Server {
 	var opts = []grpc.ServerOption{
 		grpc.Middleware(
 			recovery.Recovery(),
```

## internal/server/http.go (+1 -1)

```diff
@@ -11,7 +11,7 @@
 	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
 )
 
-func NewHTTPServer(c *conf.Server, article *service.ArticleService, logger *slog.Logger) *http.Server {
+func NewHTTPServer(c *conf.Server, article *service.Svc文章服务, logger *slog.Logger) *http.Server {
 	var opts = []http.ServerOption{
 		http.Middleware(
 			recovery.Recovery(),
```

## internal/service/article.go (+34 -34)

```diff
@@ -7,35 +7,35 @@
 	"github.com/yylego/kratos-examples/demo2kratos/internal/biz"
 )
 
-type ArticleService struct {
+type Svc文章服务 struct {
 	pb.UnimplementedArticleServiceServer
 
-	uc *biz.ArticleUsecase
+	uc *biz.Uc文章管理
 }
 
-func NewArticleService(uc *biz.ArticleUsecase) *ArticleService {
-	return &ArticleService{uc: uc}
+func NewSvc文章服务(uc *biz.Uc文章管理) *Svc文章服务 {
+	return &Svc文章服务{uc: uc}
 }
 
-func (s *ArticleService) CreateArticle(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleReply, error) {
+func (s *Svc文章服务) CreateArticle(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleReply, error) {
 	if req.Title == "" {
 		return nil, pb.ErrorBadParam("TITLE IS REQUIRED")
 	}
 	if req.StudentId <= 0 {
 		return nil, pb.ErrorBadParam("STUDENT_ID IS REQUIRED")
 	}
-	v, ebz := s.uc.CreateArticle(ctx, &biz.Article{
-		Title:     req.Title,
-		Content:   req.Content,
-		StudentID: req.StudentId,
+	v, ebz := s.uc.Xqt创建文章(ctx, &biz.Req文章信息{
+		V标题:   req.Title,
+		V内容:   req.Content,
+		V学生编号: req.StudentId,
 	})
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
-	return &pb.CreateArticleReply{Article: &pb.ArticleInfo{Id: v.ID, Title: v.Title, Content: v.Content, StudentId: v.StudentID}}, nil
+	return &pb.CreateArticleReply{Article: &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号}}, nil
 }
 
-func (s *ArticleService) UpdateArticle(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleReply, error) {
+func (s *Svc文章服务) UpdateArticle(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleReply, error) {
 	if req.Id <= 0 {
 		return nil, pb.ErrorBadParam("ID IS REQUIRED")
 	}
@@ -45,58 +45,58 @@
 	if req.StudentId <= 0 {
 		return nil, pb.ErrorBadParam("STUDENT_ID IS REQUIRED")
 	}
-	v, ebz := s.uc.UpdateArticle(ctx, &biz.Article{
-		ID:        req.Id,
-		Title:     req.Title,
-		Content:   req.Content,
-		StudentID: req.StudentId,
+	v, ebz := s.uc.Xqt更新文章(ctx, &biz.Req文章信息{
+		ID:    req.Id,
+		V标题:   req.Title,
+		V内容:   req.Content,
+		V学生编号: req.StudentId,
 	})
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
-	return &pb.UpdateArticleReply{Article: &pb.ArticleInfo{Id: v.ID, Title: v.Title, Content: v.Content, StudentId: v.StudentID}}, nil
+	return &pb.UpdateArticleReply{Article: &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号}}, nil
 }
 
-func (s *ArticleService) DeleteArticle(ctx context.Context, req *pb.DeleteArticleRequest) (*pb.DeleteArticleReply, error) {
+func (s *Svc文章服务) DeleteArticle(ctx context.Context, req *pb.DeleteArticleRequest) (*pb.DeleteArticleReply, error) {
 	if req.Id <= 0 {
 		return nil, pb.ErrorBadParam("ID IS REQUIRED")
 	}
-	if ebz := s.uc.DeleteArticle(ctx, req.Id); ebz != nil {
+	if ebz := s.uc.Xqt删除文章(ctx, req.Id); ebz != nil {
 		return nil, ebz.Erk
 	}
 	return &pb.DeleteArticleReply{Success: true}, nil
 }
 
-func (s *ArticleService) GetArticle(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleReply, error) {
+func (s *Svc文章服务) GetArticle(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleReply, error) {
 	if req.Id <= 0 {
 		return nil, pb.ErrorBadParam("ID IS REQUIRED")
 	}
-	v, ebz := s.uc.GetArticle(ctx, req.Id)
+	v, ebz := s.uc.Get获取文章(ctx, req.Id)
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
-	return &pb.GetArticleReply{Article: &pb.ArticleInfo{Id: v.ID, Title: v.Title, Content: v.Content, StudentId: v.StudentID}}, nil
+	return &pb.GetArticleReply{Article: &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号}}, nil
 }
 
-func (s *ArticleService) ListArticles(ctx context.Context, req *pb.ListArticlesRequest) (*pb.ListArticlesReply, error) {
+func (s *Svc文章服务) ListArticles(ctx context.Context, req *pb.ListArticlesRequest) (*pb.ListArticlesReply, error) {
 	if req.Page < 1 {
 		return nil, pb.ErrorBadParam("PAGE MUST BE POSITIVE")
 	}
 	if req.PageSize < 1 {
 		return nil, pb.ErrorBadParam("PAGE_SIZE MUST BE POSITIVE")
 	}
-	articles, count, ebz := s.uc.ListArticles(ctx, req.Page, req.PageSize)
+	res, ebz := s.uc.Get文章列表(ctx, &biz.Req文章列表{V页码: req.Page, V每页大小: req.PageSize})
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
-	items := make([]*pb.ArticleInfo, 0, len(articles))
-	for _, v := range articles {
-		items = append(items, &pb.ArticleInfo{Id: v.ID, Title: v.Title, Content: v.Content, StudentId: v.StudentID})
+	items := make([]*pb.ArticleInfo, 0, len(res.V文章列表))
+	for _, v := range res.V文章列表 {
+		items = append(items, &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号})
 	}
-	return &pb.ListArticlesReply{Articles: items, Count: count}, nil
+	return &pb.ListArticlesReply{Articles: items, Count: res.V总数}, nil
 }
 
-func (s *ArticleService) ListStudentArticles(ctx context.Context, req *pb.ListStudentArticlesRequest) (*pb.ListArticlesReply, error) {
+func (s *Svc文章服务) ListStudentArticles(ctx context.Context, req *pb.ListStudentArticlesRequest) (*pb.ListArticlesReply, error) {
 	if req.StudentId <= 0 {
 		return nil, pb.ErrorBadParam("STUDENT_ID IS REQUIRED")
 	}
@@ -106,13 +106,13 @@
 	if req.PageSize < 1 {
 		return nil, pb.ErrorBadParam("PAGE_SIZE MUST BE POSITIVE")
 	}
-	articles, count, ebz := s.uc.ListStudentArticles(ctx, req.StudentId, req.Page, req.PageSize)
+	res, ebz := s.uc.Get学生文章列表(ctx, &biz.Req学生文章列表{V学生编号: req.StudentId, V页码: req.Page, V每页大小: req.PageSize})
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
-	items := make([]*pb.ArticleInfo, 0, len(articles))
-	for _, v := range articles {
-		items = append(items, &pb.ArticleInfo{Id: v.ID, Title: v.Title, Content: v.Content, StudentId: v.StudentID})
+	items := make([]*pb.ArticleInfo, 0, len(res.V文章列表))
+	for _, v := range res.V文章列表 {
+		items = append(items, &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号})
 	}
-	return &pb.ListArticlesReply{Articles: items, Count: count}, nil
+	return &pb.ListArticlesReply{Articles: items, Count: res.V总数}, nil
 }
```

## internal/service/service.go (+1 -1)

```diff
@@ -2,4 +2,4 @@
 
 import "github.com/google/wire"
 
-var ProviderSet = wire.NewSet(NewArticleService)
+var ProviderSet = wire.NewSet(NewSvc文章服务)
```

