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

## cmd/demo1kratos/main.go (+11 -0)

```diff
@@ -5,6 +5,7 @@
 	"log/slog"
 	"os"
 
+	"github.com/go-kratos/kratos/contrib/encoding/json/v3"
 	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
 	"github.com/go-kratos/kratos/v3"
 	"github.com/go-kratos/kratos/v3/config"
@@ -32,6 +33,16 @@
 
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
 }
 
 func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
```

## cmd/demo1kratos/wire_gen.go (+2 -6)

```diff
@@ -28,12 +28,8 @@
 	if err != nil {
 		return nil, nil, err
 	}
-	studentUsecase, err := biz.NewStudentUsecase(dataData, logger)
-	if err != nil {
-		cleanup()
-		return nil, nil, err
-	}
-	studentService := service.NewStudentService(studentUsecase)
+	uc学生管理 := biz.NewUc学生管理(dataData, logger)
+	studentService := service.NewStudentService(uc学生管理)
 	grpcServer := server.NewGRPCServer(confServer, studentService, logger)
 	httpServer := server.NewHTTPServer(confServer, studentService, logger)
 	app := newApp(logger, grpcServer, httpServer)
```

## internal/biz/biz.go (+1 -1)

```diff
@@ -2,4 +2,4 @@
 
 import "github.com/google/wire"
 
-var ProviderSet = wire.NewSet(NewStudentUsecase)
+var ProviderSet = wire.NewSet(NewUc学生管理)
```

## internal/biz/student.go (+141 -89)

```diff
@@ -2,148 +2,200 @@
 
 import (
 	"context"
-	"errors"
 	"log/slog"
 
+	"github.com/go-kratos/kratos/v3/errors"
+	"github.com/yylego/gormcnm"
+	"github.com/yylego/gormrepo"
+	"github.com/yylego/gormrepo/gormclass"
 	"github.com/yylego/kratos-ebz/ebzkratos"
 	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/models"
+	"github.com/yylego/kratos-gorm/gormkratos"
 	"github.com/yylego/must"
 	"gorm.io/gorm"
 	"gorm.io/gorm/clause"
 )
 
-// Student is the GORM type mapped to the "students" table.
-//
-// Student 是映射到 students 表的 GORM 模型
-type Student struct {
-	ID        int64  `gorm:"primaryKey;autoIncrement"`
-	Name      string `gorm:"size:128;not null"`
-	Age       int32
-	ClassName string `gorm:"size:128"`
+type Req学生信息 struct {
+	ID  int64
+	V名字 string
+	V年龄 int32
+	V班级 string
 }
 
-func (Student) TableName() string { return "students" }
-
-// The mirrored Article type behind cascade-delete lives in article.go.
-// 用于级联删除的 Article 镜像模型定义在 article.go 中。
-
-type StudentUsecase struct {
-	data *data.Data
-	slog *slog.Logger
+type Uc学生管理 struct {
+	data   *data.Data
+	repo学生 *gormrepo.Repo[models.T学生, *models.T学生Columns]
+	repo文章 *gormrepo.Repo[models.T文章, *models.T文章Columns]
+	log    *slog.Logger
 }
 
-func NewStudentUsecase(data *data.Data, logger *slog.Logger) (*StudentUsecase, error) {
-	// Share one database with the article service: keep both tables in sync here
-	// 与文章服务共用一个库：在这里把两张表都建好
-	if err := data.DB().AutoMigrate(&Student{}, &Article{}); err != nil {
-		return nil, err
+func NewUc学生管理(data *data.Data, logger *slog.Logger) *Uc学生管理 {
+	return &Uc学生管理{
+		data:   data,
+		repo学生: gormrepo.NewRepo(gormclass.Use(&models.T学生{})),
+		repo文章: gormrepo.NewRepo(gormclass.Use(&models.T文章{})),
+		log:    logger,
 	}
-	return &StudentUsecase{data: data, slog: logger}, nil
 }
 
-func (uc *StudentUsecase) CreateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
-	must.Nice(s.Name)
+func (uc *Uc学生管理) Xqt创建学生(ctx context.Context, req *Req学生信息) (*Req学生信息, *ebzkratos.Ebz) {
+	must.Nice(req.V名字)
 
-	res := &Student{Name: s.Name, Age: s.Age, ClassName: s.ClassName}
-	if err := uc.data.DB().WithContext(ctx).Create(res).Error; err != nil {
-		return nil, ebzkratos.New(pb.ErrorStudentCreateFailure("create student: %v", err))
+	db := uc.data.DB()
+
+	var v学生 *models.T学生
+
+	if erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
+		v学生 = &models.T学生{
+			V名字: req.V名字,
+			V年龄: req.V年龄,
+			V班级: req.V班级,
+		}
+		if err := uc.repo学生.With(ctx, db).Create(v学生); err != nil {
+			return pb.ErrorStudentCreateFailure("create student: %v", err)
+		}
+		return nil
+	}); err != nil {
+		if erk != nil {
+			return nil, ebzkratos.New(erk)
+		}
+		return nil, ebzkratos.New(pb.ErrorTxError("tx: %v", err))
 	}
-	uc.slog.InfoContext(ctx, "created student", "id", res.ID, "name", res.Name)
-	return res, nil
+	return &Req学生信息{
+		ID:  int64(v学生.ID),
+		V名字: v学生.V名字,
+		V年龄: v学生.V年龄,
+		V班级: v学生.V班级,
+	}, nil
 }
 
-func (uc *StudentUsecase) UpdateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
-	must.True(s.ID > 0)
-	must.Nice(s.Name)
+func (uc *Uc学生管理) Xqt更新学生(ctx context.Context, req *Req学生信息) (*Req学生信息, *ebzkratos.Ebz) {
+	must.True(req.ID > 0)
+	must.Nice(req.V名字)
 
-	res := &Student{ID: s.ID}
-	upd := uc.data.DB().WithContext(ctx).Model(res).Updates(map[string]any{
-		"name":       s.Name,
-		"age":        s.Age,
-		"class_name": s.ClassName,
-	})
-	if upd.Error != nil {
-		return nil, ebzkratos.New(pb.ErrorDbError("update student: %v", upd.Error))
+	db := uc.data.DB()
+
+	// 先确认学生存在，对齐桩子：查不到返回 StudentNotFound 而非静默成功
+	if _, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
+		return db.Where(cls.ID.Eq(uint(req.ID)))
+	}); erb != nil {
+		if erb.NotExist {
+			return nil, ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", req.ID))
+		}
+		return nil, ebzkratos.New(pb.ErrorDbError("get student: %v", erb.Cause))
 	}
-	if upd.RowsAffected == 0 {
-		return nil, ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", s.ID))
+
+	if err := uc.repo学生.With(ctx, db).UpdatesM(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
+		return db.Where(cls.ID.Eq(uint(req.ID)))
+	}, func(cls *models.T学生Columns) gormcnm.ColumnValueMap {
+		return cls.Kw(cls.V名字.Kv(req.V名字)).Kw(cls.V年龄.Kv(req.V年龄)).Kw(cls.V班级.Kv(req.V班级))
+	}); err != nil {
+		return nil, ebzkratos.New(pb.ErrorDbError("update student: %v", err))
 	}
-	if err := uc.data.DB().WithContext(ctx).First(res, s.ID).Error; err != nil {
-		return nil, ebzkratos.New(pb.ErrorDbError("reload student: %v", err))
-	}
-	return res, nil
+
+	return req, nil
 }
 
-func (uc *StudentUsecase) DeleteStudent(ctx context.Context, id int64) *ebzkratos.Ebz {
+func (uc *Uc学生管理) Xqt删除学生(ctx context.Context, id int64) *ebzkratos.Ebz {
 	must.True(id > 0)
 
-	// Atomic, race-safe cascade delete, in one transaction:
-	//   1. lock the student row (FOR UPDATE) so no article can target
-	//      this student meanwhile — CreateArticle takes a conflicting FOR SHARE
-	//      lock on the same row, so the two operations serialize;
-	//   2. delete the student's articles (children first);
-	//   3. delete the student (parent last).
-	// 原子且并发安全的级联删除，全部在一个事务里完成：
-	//   ① 用 FOR UPDATE 锁住学生行，删除期间不允许给该学生并发新建文章——CreateArticle
-	//      会对同一行加互斥的 FOR SHARE 锁，二者因此串行化；
+	db := uc.data.DB()
+
+	// 翻译桩子的原子级联删除，全在一个事务里：
+	//   ① FOR UPDATE 锁住学生行，删除期间挡住并发新建文章（对方持互斥的 FOR SHARE 锁）；
 	//   ② 先删该学生名下的文章（子表在前）；
 	//   ③ 再删学生本身（父表在后）。
-	var notFound bool
-	var removedArticles int64
-	err := uc.data.DB().WithContext(ctx).Transaction(func(db *gorm.DB) error {
-		var s Student
-		if err := db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).First(&s, id).Error; err != nil {
-			if errors.Is(err, gorm.ErrRecordNotFound) {
-				notFound = true
+	var b未找到 bool
+	if erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
+		if _, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
+			return db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).Where(cls.ID.Eq(uint(id)))
+		}); erb != nil {
+			if erb.NotExist {
+				b未找到 = true
 				return nil
 			}
-			return err
+			return pb.ErrorDbError("get student: %v", erb.Cause)
 		}
-		del := db.Where("student_id = ?", id).Delete(&Article{})
-		if del.Error != nil {
-			return del.Error
+		if err := uc.repo文章.With(ctx, db).DeleteW(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
+			return db.Where(cls.V学生编号.Eq(id))
+		}); err != nil {
+			return pb.ErrorDbError("delete articles: %v", err)
 		}
-		removedArticles = del.RowsAffected
-		return db.Delete(&Student{}, id).Error
-	})
-	if err != nil {
+		if err := uc.repo学生.With(ctx, db).DeleteW(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
+			return db.Where(cls.ID.Eq(uint(id)))
+		}); err != nil {
+			return pb.ErrorDbError("delete student: %v", err)
+		}
+		return nil
+	}); err != nil {
+		if erk != nil {
+			return ebzkratos.New(erk)
+		}
 		return ebzkratos.New(pb.ErrorTxError("delete student with articles: %v", err))
 	}
-	if notFound {
+	if b未找到 {
 		return ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", id))
 	}
-	uc.slog.InfoContext(ctx, "deleted student and cascaded articles", "student_id", id, "articles_removed", removedArticles)
 	return nil
 }
 
-func (uc *StudentUsecase) GetStudent(ctx context.Context, id int64) (*Student, *ebzkratos.Ebz) {
+func (uc *Uc学生管理) Get获取学生(ctx context.Context, id int64) (*Req学生信息, *ebzkratos.Ebz) {
 	must.True(id > 0)
 
-	res := &Student{}
-	if err := uc.data.DB().WithContext(ctx).First(res, id).Error; err != nil {
-		if errors.Is(err, gorm.ErrRecordNotFound) {
+	db := uc.data.DB()
+
+	v学生, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
+		return db.Where(cls.ID.Eq(uint(id)))
+	})
+	if erb != nil {
+		if erb.NotExist {
 			return nil, ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", id))
 		}
-		return nil, ebzkratos.New(pb.ErrorDbError("get student: %v", err))
+		return nil, ebzkratos.New(pb.ErrorDbError("get student: %v", erb.Cause))
 	}
-	return res, nil
+
+	return &Req学生信息{
+		ID:  int64(v学生.ID),
+		V名字: v学生.V名字,
+		V年龄: v学生.V年龄,
+		V班级: v学生.V班级,
+	}, nil
 }
 
-func (uc *StudentUsecase) ListStudents(ctx context.Context, page int32, pageSize int32) ([]*Student, int32, *ebzkratos.Ebz) {
+func (uc *Uc学生管理) Get学生列表(ctx context.Context, page int32, pageSize int32) ([]*Req学生信息, int32, *ebzkratos.Ebz) {
 	must.True(page >= 1)
 	must.True(pageSize >= 1)
 
-	db := uc.data.DB().WithContext(ctx)
+	db := uc.data.DB()
 
-	var count int64
-	if err := db.Model(&Student{}).Count(&count).Error; err != nil {
-		return nil, 0, ebzkratos.New(pb.ErrorDbError("count students: %v", err))
+	// gormrepo FindPageAndCount 一次拿到当页数据和总行数，对齐桩子的分页+计数。
+	v学生们, total, err := uc.repo学生.With(ctx, db).FindPageAndCount(
+		func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
+			return db
+		},
+		func(cls *models.T学生Columns) gormcnm.OrderByBottle {
+			return cls.ID.Ob("asc")
+		},
+		&gormrepo.Pagination{
+			Offset: int((page - 1) * pageSize),
+			Limit:  int(pageSize),
+		},
+	)
+	if err != nil {
+		return nil, 0, ebzkratos.New(pb.ErrorDbError("list students: %v", err))
 	}
 
-	var items []*Student
-	if err := db.Order("id").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&items).Error; err != nil {
-		return nil, 0, ebzkratos.New(pb.ErrorDbError("list students: %v", err))
+	a学生列表 := make([]*Req学生信息, 0, len(v学生们))
+	for _, v := range v学生们 {
+		a学生列表 = append(a学生列表, &Req学生信息{
+			ID:  int64(v.ID),
+			V名字: v.V名字,
+			V年龄: v.V年龄,
+			V班级: v.V班级,
+		})
 	}
-	return items, int32(count), nil
+	return a学生列表, int32(total), nil
 }
```

## internal/data/data.go (+9 -8)

```diff
@@ -5,6 +5,7 @@
 
 	"github.com/google/wire"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/models"
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
+	must.Done(db.AutoMigrate(&models.T学生{}, &models.T文章{}))
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
+	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/models"
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
+		&models.T学生{},
+		&models.T文章{},
+	}
+
+	// Configure generation options with latest best practices
+	options := gormcngen.NewOptions().
+		WithColumnClassExportable(true). // Generate exportable column class names like T学生Columns
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

## internal/service/student.go (+20 -20)

```diff
@@ -10,10 +10,10 @@
 type StudentService struct {
 	pb.UnimplementedStudentServiceServer
 
-	uc *biz.StudentUsecase
+	uc *biz.Uc学生管理
 }
 
-func NewStudentService(uc *biz.StudentUsecase) *StudentService {
+func NewStudentService(uc *biz.Uc学生管理) *StudentService {
 	return &StudentService{uc: uc}
 }
 
@@ -21,15 +21,15 @@
 	if req.Name == "" {
 		return nil, pb.ErrorBadParam("NAME IS REQUIRED")
 	}
-	v, ebz := s.uc.CreateStudent(ctx, &biz.Student{
-		Name:      req.Name,
-		Age:       req.Age,
-		ClassName: req.ClassName,
+	v, ebz := s.uc.Xqt创建学生(ctx, &biz.Req学生信息{
+		V名字: req.Name,
+		V年龄: req.Age,
+		V班级: req.ClassName,
 	})
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
-	return &pb.CreateStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.Name, Age: v.Age, ClassName: v.ClassName}}, nil
+	return &pb.CreateStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.V名字, Age: v.V年龄, ClassName: v.V班级}}, nil
 }
 
 func (s *StudentService) UpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) (*pb.UpdateStudentReply, error) {
@@ -39,23 +39,23 @@
 	if req.Name == "" {
 		return nil, pb.ErrorBadParam("NAME IS REQUIRED")
 	}
-	v, ebz := s.uc.UpdateStudent(ctx, &biz.Student{
-		ID:        req.Id,
-		Name:      req.Name,
-		Age:       req.Age,
-		ClassName: req.ClassName,
+	v, ebz := s.uc.Xqt更新学生(ctx, &biz.Req学生信息{
+		ID:  req.Id,
+		V名字: req.Name,
+		V年龄: req.Age,
+		V班级: req.ClassName,
 	})
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
-	return &pb.UpdateStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.Name, Age: v.Age, ClassName: v.ClassName}}, nil
+	return &pb.UpdateStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.V名字, Age: v.V年龄, ClassName: v.V班级}}, nil
 }
 
 func (s *StudentService) DeleteStudent(ctx context.Context, req *pb.DeleteStudentRequest) (*pb.DeleteStudentReply, error) {
 	if req.Id <= 0 {
 		return nil, pb.ErrorBadParam("ID IS REQUIRED")
 	}
-	if ebz := s.uc.DeleteStudent(ctx, req.Id); ebz != nil {
+	if ebz := s.uc.Xqt删除学生(ctx, req.Id); ebz != nil {
 		return nil, ebz.Erk
 	}
 	return &pb.DeleteStudentReply{Success: true}, nil
@@ -65,11 +65,11 @@
 	if req.Id <= 0 {
 		return nil, pb.ErrorBadParam("ID IS REQUIRED")
 	}
-	v, ebz := s.uc.GetStudent(ctx, req.Id)
+	v, ebz := s.uc.Get获取学生(ctx, req.Id)
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
-	return &pb.GetStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.Name, Age: v.Age, ClassName: v.ClassName}}, nil
+	return &pb.GetStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.V名字, Age: v.V年龄, ClassName: v.V班级}}, nil
 }
 
 func (s *StudentService) ListStudents(ctx context.Context, req *pb.ListStudentsRequest) (*pb.ListStudentsReply, error) {
@@ -79,13 +79,13 @@
 	if req.PageSize < 1 {
 		return nil, pb.ErrorBadParam("PAGE_SIZE MUST BE POSITIVE")
 	}
-	students, count, ebz := s.uc.ListStudents(ctx, req.Page, req.PageSize)
+	a学生列表, count, ebz := s.uc.Get学生列表(ctx, req.Page, req.PageSize)
 	if ebz != nil {
 		return nil, ebz.Erk
 	}
-	items := make([]*pb.StudentInfo, 0, len(students))
-	for _, v := range students {
-		items = append(items, &pb.StudentInfo{Id: v.ID, Name: v.Name, Age: v.Age, ClassName: v.ClassName})
+	items := make([]*pb.StudentInfo, 0, len(a学生列表))
+	for _, v := range a学生列表 {
+		items = append(items, &pb.StudentInfo{Id: v.ID, Name: v.V名字, Age: v.V年龄, ClassName: v.V班级})
 	}
 	return &pb.ListStudentsReply{Students: items, Count: count}, nil
 }
```

