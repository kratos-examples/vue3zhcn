# Changes

Code differences compared to source project.

## Makefile (+14 -0)

```diff
@@ -36,6 +36,20 @@
 	make config
 	make generate
 
+# generate TypeScript gRPC clients from proto files (via buf)
+# go install github.com/yylego/kratos-vue3/cmd/vue3kratos@latest
+web_api_grpc_ts:
+	mkdir -p ./bin/web_api_grpc_ts.out
+	buf generate --template buf.gen.ts.yaml --include-imports
+
+# convert gRPC clients to HTTP clients
+web_api_grpc_to_http:
+	vue3kratos gen-grpc-via-http-in-root --grpc-ts-root=./bin/web_api_grpc_ts.out
+
+# cleanup generated TypeScript files
+web_api_cleanup:
+	rm -rf ./bin/web_api_grpc_ts.out
+
 # show help
 help:
 	@echo ''
```

## buf.gen.ts.yaml (+10 -0)

```diff
@@ -0,0 +1,10 @@
+version: v2
+inputs:
+  - directory: api
+plugins:
+  - local: protoc-gen-ts
+    out: bin/web_api_grpc_ts.out
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
@@ -32,6 +33,17 @@
 
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

## cmd/demo2kratos/wire_gen.go (+2 -6)

```diff
@@ -28,12 +28,8 @@
 	if err != nil {
 		return nil, nil, err
 	}
-	articleUsecase, err := biz.NewArticleUsecase(dataData, logger)
-	if err != nil {
-		cleanup()
-		return nil, nil, err
-	}
-	articleService := service.NewArticleService(articleUsecase)
+	uc文章管理 := biz.NewUc文章管理(dataData, logger)
+	articleService := service.NewArticleService(uc文章管理)
 	grpcServer := server.NewGRPCServer(confServer, articleService, logger)
 	httpServer := server.NewHTTPServer(confServer, articleService, logger)
 	app := newApp(logger, grpcServer, httpServer)
```

## internal/biz/article.go (+126 -143)

```diff
@@ -2,195 +2,178 @@
 
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
+	log    *slog.Logger
 }
 
-func NewArticleUsecase(data *data.Data, logger *slog.Logger) (*ArticleUsecase, error) {
-	// Migrate the owned table plus the mirrored students table (needed in the
-	// existence check); both services share one database
-	// 建好本服务拥有的 articles 表，外加镜像的 students 表（供存在性校验用）
-	if err := data.DB().AutoMigrate(&Article{}, &Student{}); err != nil {
-		return nil, err
+func NewUc文章管理(data *data.Data, logger *slog.Logger) *Uc文章管理 {
+	return &Uc文章管理{
+		data:   data,
+		repo文章: gormrepo.NewRepo(gormclass.Use(&models.T文章{})),
+		repo学生: gormrepo.NewRepo(gormclass.Use(&models.T学生{})),
+		log:    logger,
 	}
-	return &ArticleUsecase{data: data, slog: logger}, nil
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
+			return db.Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).Where(cls.ID.Eq(uint(req.V学生编号)))
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
-	return res, nil
+	return &Req文章信息{
+		ID:    int64(v文章.ID),
+		V标题:   v文章.V标题,
+		V内容:   v文章.V内容,
+		V学生编号: v文章.V学生编号,
+	}, nil
 }
 
-func (uc *ArticleUsecase) UpdateArticle(ctx context.Context, a *Article) (*Article, *ebzkratos.Ebz) {
-	must.True(a.ID > 0)
-	must.Nice(a.Title)
-	must.True(a.StudentID > 0)
+func (uc *Uc文章管理) Xqt更新文章(ctx context.Context, req *Req文章信息) (*Req文章信息, *ebzkratos.Ebz) {
+	must.True(req.ID > 0)
+	must.Nice(req.V标题)
 
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
-			}
-			return err
+	db := uc.data.DB()
+
+	// 先确认文章存在，对齐桩子：查不到返回 ArticleNotFound 而非静默成功
+	if _, erb := uc.repo文章.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+		return db.Where(cls.ID.Eq(uint(req.ID)))
+	}); erb != nil {
+		if erb.NotExist {
+			return nil, ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", req.ID))
 		}
-		upd := db.Model(res).Updates(map[string]any{
-			"title":      a.Title,
-			"content":    a.Content,
-			"student_id": a.StudentID,
-		})
-		if upd.Error != nil {
-			return upd.Error
-		}
-		if upd.RowsAffected == 0 {
-			articleMissing = true
-			return nil
-		}
-		return db.First(res, a.ID).Error
-	})
-	if err != nil {
+		return nil, ebzkratos.New(pb.ErrorDbError("get article: %v", erb.Cause))
+	}
+
+	if err := uc.repo文章.With(ctx, db).UpdatesM(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+		return db.Where(cls.ID.Eq(uint(req.ID)))
+	}, func(cls *models.T文章Columns) gormcnm.ColumnValueMap {
+		return cls.Kw(cls.V标题.Kv(req.V标题)).Kw(cls.V内容.Kv(req.V内容)).Kw(cls.V学生编号.Kv(req.V学生编号))
+	}); err != nil {
 		return nil, ebzkratos.New(pb.ErrorDbError("update article: %v", err))
 	}
-	if studentMissing {
-		return nil, ebzkratos.New(pb.ErrorBadParam("student %d does not exist", a.StudentID))
-	}
-	if articleMissing {
-		return nil, ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", a.ID))
-	}
-	return res, nil
+
+	return req, nil
 }
 
-func (uc *ArticleUsecase) DeleteArticle(ctx context.Context, id int64) *ebzkratos.Ebz {
+func (uc *Uc文章管理) Xqt删除文章(ctx context.Context, id int64) *ebzkratos.Ebz {
 	must.True(id > 0)
 
-	del := uc.data.DB().WithContext(ctx).Delete(&Article{}, id)
-	if del.Error != nil {
-		return ebzkratos.New(pb.ErrorDbError("delete article: %v", del.Error))
+	db := uc.data.DB()
+
+	// 先确认文章存在，对齐桩子：查不到返回 ArticleNotFound
+	if _, erb := uc.repo文章.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+		return db.Where(cls.ID.Eq(uint(id)))
+	}); erb != nil {
+		if erb.NotExist {
+			return ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", id))
+		}
+		return ebzkratos.New(pb.ErrorDbError("get article: %v", erb.Cause))
 	}
-	if del.RowsAffected == 0 {
-		return ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", id))
+
+	if err := uc.repo文章.With(ctx, db).DeleteW(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+		return db.Where(cls.ID.Eq(uint(id)))
+	}); err != nil {
+		return ebzkratos.New(pb.ErrorDbError("delete article: %v", err))
 	}
-	uc.slog.InfoContext(ctx, "deleted article", "id", id)
 	return nil
 }
 
-func (uc *ArticleUsecase) GetArticle(ctx context.Context, id int64) (*Article, *ebzkratos.Ebz) {
+func (uc *Uc文章管理) Get获取文章(ctx context.Context, id int64) (*Req文章信息, *ebzkratos.Ebz) {
 	must.True(id > 0)
 
-	res := &Article{}
-	if err := uc.data.DB().WithContext(ctx).First(res, id).Error; err != nil {
-		if errors.Is(err, gorm.ErrRecordNotFound) {
+	db := uc.data.DB()
+
+	v文章, erb := uc.repo文章.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+		return db.Where(cls.ID.Eq(uint(id)))
+	})
+	if erb != nil {
+		if erb.NotExist {
 			return nil, ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", id))
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
-	if page < 1 {
-		page = 1
-	}
-	if pageSize < 1 {
-		pageSize = 10
-	}
+func (uc *Uc文章管理) Get文章列表(ctx context.Context, page int32, pageSize int32) ([]*Req文章信息, int32, *ebzkratos.Ebz) {
+	db := uc.data.DB()
 
-	db := uc.data.DB().WithContext(ctx)
-
-	var total int64
-	if err := db.Model(&Article{}).Count(&total).Error; err != nil {
-		return nil, 0, ebzkratos.New(pb.ErrorDbError("count articles: %v", err))
-	}
-
-	var items []*Article
-	if err := db.Order("id").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&items).Error; err != nil {
+	v文章们, err := uc.repo文章.With(ctx, db).Find(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+		return db.Order(cls.ID.Ob("DESC").Ox())
+	})
+	if err != nil {
 		return nil, 0, ebzkratos.New(pb.ErrorDbError("list articles: %v", err))
 	}
-	return items, int32(total), nil
-}
 
-// ListStudentArticles returns one student's articles, one page at a time. The
-// student↔article relationship gets its own endpoint instead of overloading
-// ListArticles with an extra flag.
-//
-// ListStudentArticles 分页返回某个学生的文章。学生↔文章这层关系单独开一个接口，
-// 而不是往 ListArticles 上塞过滤参数。
-func (uc *ArticleUsecase) ListStudentArticles(ctx context.Context, studentID int64, page int32, pageSize int32) ([]*Article, int32, *ebzkratos.Ebz) {
-	must.True(studentID > 0)
-	if page < 1 {
-		page = 1
+	a文章列表 := make([]*Req文章信息, 0, len(v文章们))
+	for _, v := range v文章们 {
+		a文章列表 = append(a文章列表, &Req文章信息{
+			ID:    int64(v.ID),
+			V标题:   v.V标题,
+			V内容:   v.V内容,
+			V学生编号: v.V学生编号,
+		})
 	}
-	if pageSize < 1 {
-		pageSize = 10
-	}
-
-	db := uc.data.DB().WithContext(ctx)
-
-	var total int64
-	if err := db.Model(&Article{}).Where("student_id = ?", studentID).Count(&total).Error; err != nil {
-		return nil, 0, ebzkratos.New(pb.ErrorDbError("count student articles: %v", err))
-	}
-
-	var items []*Article
-	if err := db.Where("student_id = ?", studentID).Order("id").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&items).Error; err != nil {
-		return nil, 0, ebzkratos.New(pb.ErrorDbError("list student articles: %v", err))
-	}
-	return items, int32(total), nil
+	return a文章列表, int32(len(a文章列表)), nil
 }
```

## internal/biz/biz.go (+1 -1)

```diff
@@ -2,4 +2,4 @@
 
 import "github.com/google/wire"
 
-var ProviderSet = wire.NewSet(NewArticleUsecase)
+var ProviderSet = wire.NewSet(NewUc文章管理)
```

## internal/data/data.go (+9 -8)

```diff
@@ -5,6 +5,7 @@
 
 	"github.com/google/wire"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
+	"github.com/yylego/kratos-examples/demo2kratos/internal/pkg/models"
 	"github.com/yylego/must"
 	"github.com/yylego/rese"
 	"gorm.io/driver/postgres"
@@ -17,19 +18,19 @@
 	db *gorm.DB
 }
 
-// DB exposes the underlying gorm handle so the biz code can run true queries.
-//
-// DB 暴露底层 gorm 句柄，供 biz 层执行真实的数据库读写
-func (d *Data) DB() *gorm.DB {
-	return d.db
-}
-
 func NewData(c *conf.Data, logger *slog.Logger) (*Data, func(), error) {
 	must.Same(c.Database.Driver, "postgres")
 	db := rese.P1(gorm.Open(postgres.Open(c.Database.Source), &gorm.Config{}))
+
+	must.Done(db.AutoMigrate(&models.T文章{}, &models.T学生{}))
+
 	cleanup := func() {
 		logger.Info("closing the data resources")
-		_ = rese.P1(db.DB()).Close()
+		must.Done(rese.P1(db.DB()).Close())
 	}
 	return &Data{db: db}, cleanup, nil
+}
+
+func (d *Data) DB() *gorm.DB {
+	return d.db
 }
```

## internal/pkg/models/article.go (+14 -0)

```diff
@@ -0,0 +1,14 @@
+package models
+
+import "gorm.io/gorm"
+
+type T文章 struct {
+	gorm.Model
+	V标题   string `gorm:"column:title;type:varchar(255)" cnm:"V标题"`
+	V内容   string `gorm:"column:content;type:text" cnm:"V内容"`
+	V学生编号 int64  `gorm:"column:student_id;type:bigint" cnm:"V学生编号"`
+}
+
+func (*T文章) TableName() string {
+	return "articles"
+}
```

## internal/pkg/models/gormcnm.gen.go (+71 -0)

```diff
@@ -0,0 +1,71 @@
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
+	"time"
+
+	"github.com/yylego/gormcnm"
+	"gorm.io/gorm"
+)
+
+func (c *T文章) Columns() *T文章Columns {
+	return &T文章Columns{
+		// Auto-generated: column names and types mapping. DO NOT EDIT. // 自动生成：列名和类型映射。请勿编辑。
+		ID:        gormcnm.Cnm(c.ID, "id"),
+		CreatedAt: gormcnm.Cnm(c.CreatedAt, "created_at"),
+		UpdatedAt: gormcnm.Cnm(c.UpdatedAt, "updated_at"),
+		DeletedAt: gormcnm.Cnm(c.DeletedAt, "deleted_at"),
+		V标题:       gormcnm.Cnm(c.V标题, "title"),
+		V内容:       gormcnm.Cnm(c.V内容, "content"),
+		V学生编号:     gormcnm.Cnm(c.V学生编号, "student_id"),
+	}
+}
+
+type T文章Columns struct {
+	// Auto-generated: embedding operation functions to make it simple to use. DO NOT EDIT. // 自动生成：嵌入操作函数便于使用。请勿编辑。
+	gormcnm.ColumnOperationClass
+	// Auto-generated: column names and types in database table. DO NOT EDIT. // 自动生成：数据库表的列名和类型。请勿编辑。
+	ID        gormcnm.ColumnName[uint]
+	CreatedAt gormcnm.ColumnName[time.Time]
+	UpdatedAt gormcnm.ColumnName[time.Time]
+	DeletedAt gormcnm.ColumnName[gorm.DeletedAt]
+	V标题       gormcnm.ColumnName[string]
+	V内容       gormcnm.ColumnName[string]
+	V学生编号     gormcnm.ColumnName[int64]
+}
+
+func (c *T学生) Columns() *T学生Columns {
+	return &T学生Columns{
+		// Auto-generated: column names and types mapping. DO NOT EDIT. // 自动生成：列名和类型映射。请勿编辑。
+		ID:        gormcnm.Cnm(c.ID, "id"),
+		CreatedAt: gormcnm.Cnm(c.CreatedAt, "created_at"),
+		UpdatedAt: gormcnm.Cnm(c.UpdatedAt, "updated_at"),
+		DeletedAt: gormcnm.Cnm(c.DeletedAt, "deleted_at"),
+		V名字:       gormcnm.Cnm(c.V名字, "name"),
+		V年龄:       gormcnm.Cnm(c.V年龄, "age"),
+		V班级:       gormcnm.Cnm(c.V班级, "class_name"),
+	}
+}
+
+type T学生Columns struct {
+	// Auto-generated: embedding operation functions to make it simple to use. DO NOT EDIT. // 自动生成：嵌入操作函数便于使用。请勿编辑。
+	gormcnm.ColumnOperationClass
+	// Auto-generated: column names and types in database table. DO NOT EDIT. // 自动生成：数据库表的列名和类型。请勿编辑。
+	ID        gormcnm.ColumnName[uint]
+	CreatedAt gormcnm.ColumnName[time.Time]
+	UpdatedAt gormcnm.ColumnName[time.Time]
+	DeletedAt gormcnm.ColumnName[gorm.DeletedAt]
+	V名字       gormcnm.ColumnName[string]
+	V年龄       gormcnm.ColumnName[int32]
+	V班级       gormcnm.ColumnName[string]
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

## internal/pkg/models/student.go (+14 -0)

```diff
@@ -0,0 +1,14 @@
+package models
+
+import "gorm.io/gorm"
+
+type T学生 struct {
+	gorm.Model
+	V名字 string `gorm:"column:name;type:varchar(255)" cnm:"V名字"`
+	V年龄 int32  `gorm:"column:age;type:int" cnm:"V年龄"`
+	V班级 string `gorm:"column:class_name;type:varchar(255)" cnm:"V班级"`
+}
+
+func (*T学生) TableName() string {
+	return "students"
+}
```

## internal/service/article.go (+20 -41)

```diff
@@ -10,10 +10,10 @@
 type ArticleService struct {
 	pb.UnimplementedArticleServiceServer
 
-	uc *biz.ArticleUsecase
+	uc *biz.Uc文章管理
 }
 
-func NewArticleService(uc *biz.ArticleUsecase) *ArticleService {
+func NewArticleService(uc *biz.Uc文章管理) *ArticleService {
 	return &ArticleService{uc: uc}
 }
 
@@ -21,18 +21,15 @@
 	if req.Title == "" {
 		return nil, pb.ErrorBadParam("TITLE IS REQUIRED")
 	}
-	if req.StudentId <= 0 {
-		return nil, pb.ErrorBadParam("STUDENT_ID IS REQUIRED")
-	}
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
 
 func (s *ArticleService) UpdateArticle(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleReply, error) {
@@ -42,26 +39,23 @@
 	if req.Title == "" {
 		return nil, pb.ErrorBadParam("TITLE IS REQUIRED")
 	}
-	if req.StudentId <= 0 {
-		return nil, pb.ErrorBadParam("STUDENT_ID IS REQUIRED")
-	}
-	v, ebz := s.uc.UpdateArticle(ctx, &biz.Article{
-		ID:        req.Id,
-		Title:     req.Title,
-		Content:   req.Content,
-		StudentID: req.StudentId,
+	v, ebz := s.uc.Xqt更新文章(ctx, &biz.Req文章信息{
+		ID:   req.Id,
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
 
 func (s *ArticleService) DeleteArticle(ctx context.Context, req *pb.DeleteArticleRequest) (*pb.DeleteArticleReply, error) {
 	if req.Id <= 0 {
 		return nil, pb.ErrorBadParam("ID IS REQUIRED")
 	}
-	if ebz := s.uc.DeleteArticle(ctx, req.Id); ebz != nil {
+	if ebz := s.uc.Xqt删除文章(ctx, req.Id); ebz != nil {
 		return nil, ebz.Erk
 	}
 	return &pb.DeleteArticleReply{Success: true}, nil
@@ -71,36 +65,21 @@
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
 
 func (s *ArticleService) ListArticles(ctx context.Context, req *pb.ListArticlesRequest) (*pb.ListArticlesReply, error) {
-	articles, count, ebz := s.uc.ListArticles(ctx, req.Page, req.PageSize)
+	a文章列表, count, ebz := s.uc.Get文章列表(ctx, req.Page, req.PageSize)
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
-	items := make([]*pb.ArticleInfo, 0, len(articles))
-	for _, v := range articles {
-		items = append(items, &pb.ArticleInfo{Id: v.ID, Title: v.Title, Content: v.Content, StudentId: v.StudentID})
-	}
-	return &pb.ListArticlesReply{Articles: items, Count: count}, nil
-}
-
-func (s *ArticleService) ListStudentArticles(ctx context.Context, req *pb.ListStudentArticlesRequest) (*pb.ListArticlesReply, error) {
-	if req.StudentId <= 0 {
-		return nil, pb.ErrorBadParam("STUDENT_ID IS REQUIRED")
-	}
-	articles, count, ebz := s.uc.ListStudentArticles(ctx, req.StudentId, req.Page, req.PageSize)
-	if ebz != nil {
-		return nil, ebz.Erk
-	}
-	items := make([]*pb.ArticleInfo, 0, len(articles))
-	for _, v := range articles {
-		items = append(items, &pb.ArticleInfo{Id: v.ID, Title: v.Title, Content: v.Content, StudentId: v.StudentID})
+	items := make([]*pb.ArticleInfo, 0, len(a文章列表))
+	for _, v := range a文章列表 {
+		items = append(items, &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号})
 	}
 	return &pb.ListArticlesReply{Articles: items, Count: count}, nil
 }
```

