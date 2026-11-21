package biz

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/yylego/gormcnm"
	"github.com/yylego/gormrepo"
	"github.com/yylego/gormrepo/gormclass"
	"github.com/yylego/kratos-ebz/ebzkratos"
	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
	"github.com/yylego/kratos-examples/demo1kratos/internal/data"
	"github.com/yylego/kratos-examples/demo1kratos/internal/pkg/models"
	"github.com/yylego/kratos-gorm/gormkratos"
	"github.com/yylego/must"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Req学生信息 struct {
	ID  int64
	V名字 string
	V年龄 int32
	V班级 string
}

type Uc学生管理 struct {
	data   *data.Data
	repo学生 *gormrepo.Repo[models.T学生, *models.T学生Columns]
	repo文章 *gormrepo.Repo[models.T文章, *models.T文章Columns]
	slog   *slog.Logger
}

func NewUc学生管理(data *data.Data, logger *slog.Logger) (*Uc学生管理, error) {
	// 与文章服务共用一个库：在这里把两张表都建好。
	if err := data.DB().AutoMigrate(&models.T学生{}, &models.T文章{}); err != nil {
		return nil, err
	}
	return &Uc学生管理{
		data:   data,
		repo学生: gormrepo.NewRepo(gormclass.Use(&models.T学生{})),
		repo文章: gormrepo.NewRepo(gormclass.Use(&models.T文章{})),
		slog:   logger,
	}, nil
}

func (uc *Uc学生管理) Xqt创建学生(ctx context.Context, req *Req学生信息) (*Req学生信息, *ebzkratos.Ebz) {
	must.Nice(req.V名字)

	db := uc.data.DB()

	// 忠实翻译桩子：单条 Create（桩子创建学生不用事务），失败即 StudentCreateFailure。
	v学生 := &models.T学生{
		V名字: req.V名字,
		V年龄: req.V年龄,
		V班级: req.V班级,
	}
	if err := uc.repo学生.With(ctx, db).Create(v学生); err != nil {
		return nil, ebzkratos.New(pb.ErrorStudentCreateFailure("create student: %v", err))
	}
	res := &Req学生信息{
		ID:  int64(v学生.ID),
		V名字: v学生.V名字,
		V年龄: v学生.V年龄,
		V班级: v学生.V班级,
	}
	uc.slog.InfoContext(ctx, "created student", "id", res.ID, "name", res.V名字)
	return res, nil
}

func (uc *Uc学生管理) Xqt更新学生(ctx context.Context, req *Req学生信息) (*Req学生信息, *ebzkratos.Ebz) {
	must.True(req.ID > 0)
	must.Nice(req.V名字)

	db := uc.data.DB()

	// 忠实翻译桩子：单条 UpdatesM 更新后靠 RowsAffected 判不存在（0 行→StudentNotFound），与桩子一致。
	upd := uc.repo学生.Wrap(ctx, db).UpdatesM(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(req.ID))
	}, func(cls *models.T学生Columns) gormcnm.ColumnValueMap {
		return cls.Kw(cls.V名字.Kv(req.V名字)).Kw(cls.V年龄.Kv(req.V年龄)).Kw(cls.V班级.Kv(req.V班级))
	})
	if upd.Error != nil {
		return nil, ebzkratos.New(pb.ErrorDbError("update student: %v", upd.Error))
	}
	if upd.RowsAffected == 0 {
		return nil, ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", req.ID))
	}

	// 重新载入并返回当前行，与桩子一致。
	v学生, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(req.ID))
	})
	if erb != nil {
		return nil, ebzkratos.New(pb.ErrorDbError("reload student: %v", erb.Cause))
	}
	return &Req学生信息{
		ID:  int64(v学生.ID),
		V名字: v学生.V名字,
		V年龄: v学生.V年龄,
		V班级: v学生.V班级,
	}, nil
}

func (uc *Uc学生管理) Xqt删除学生(ctx context.Context, n编号 int64) *ebzkratos.Ebz {
	must.True(n编号 > 0)

	db := uc.data.DB()

	// 翻译桩子的原子级联删除，全在一个事务里：
	//   ① FOR UPDATE 锁住学生行，删除期间挡住并发新建文章（对方持互斥的 FOR SHARE 锁）；
	//   ② 先删该学生名下的文章（子表在前）；
	//   ③ 再删学生本身（父表在后）。
	var b未找到 bool
	var removedArticles int64
	if erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		if _, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
			return db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).Where(cls.ID.Eq(n编号))
		}); erb != nil {
			if erb.NotExist {
				b未找到 = true
				return nil
			}
			return pb.ErrorDbError("get student: %v", erb.Cause)
		}
		del := uc.repo文章.Wrap(ctx, db).DeleteW(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
			return db.Where(cls.V学生编号.Eq(n编号))
		})
		if del.Error != nil {
			return pb.ErrorDbError("delete articles: %v", del.Error)
		}
		removedArticles = del.RowsAffected
		if err := uc.repo学生.With(ctx, db).DeleteW(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
			return db.Where(cls.ID.Eq(n编号))
		}); err != nil {
			return pb.ErrorDbError("delete student: %v", err)
		}
		return nil
	}); err != nil {
		if erk != nil {
			return ebzkratos.New(erk)
		}
		return ebzkratos.New(pb.ErrorTxError("delete student with articles: %v", err))
	}
	if b未找到 {
		return ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", n编号))
	}
	uc.slog.InfoContext(ctx, "deleted student and cascaded articles", "student_id", n编号, "articles_removed", removedArticles)
	return nil
}

func (uc *Uc学生管理) Get获取学生(ctx context.Context, n编号 int64) (*Req学生信息, *ebzkratos.Ebz) {
	must.True(n编号 > 0)

	db := uc.data.DB()

	v学生, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(n编号))
	})
	if erb != nil {
		if erb.NotExist {
			return nil, ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", n编号))
		}
		return nil, ebzkratos.New(pb.ErrorDbError("get student: %v", erb.Cause))
	}

	return &Req学生信息{
		ID:  int64(v学生.ID),
		V名字: v学生.V名字,
		V年龄: v学生.V年龄,
		V班级: v学生.V班级,
	}, nil
}

// Req学生列表 包裹「查询学生列表」的入参（分页），Res学生列表 包裹返回（当页数据+总数）。
type Req学生列表 struct {
	V页码   int32
	V每页大小 int32
}

type Res学生列表 struct {
	V学生列表 []*Req学生信息
	V总数   int32
}

func (uc *Uc学生管理) Get学生列表(ctx context.Context, req *Req学生列表) (*Res学生列表, *ebzkratos.Ebz) {
	must.True(req.V页码 >= 1)
	must.True(req.V每页大小 >= 1)

	db := uc.data.DB()

	// gormrepo FindPageAndCount 一次拿到当页数据和总行数，对齐桩子的分页+计数。
	v学生们, total, err := uc.repo学生.With(ctx, db).FindPageAndCount(
		func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
			return db
		},
		func(cls *models.T学生Columns) gormcnm.OrderByBottle {
			return cls.ID.Ob("asc")
		},
		&gormrepo.Pagination{
			Offset: int((req.V页码 - 1) * req.V每页大小),
			Limit:  int(req.V每页大小),
		},
	)
	if err != nil {
		return nil, ebzkratos.New(pb.ErrorDbError("list students: %v", err))
	}

	a学生列表 := make([]*Req学生信息, 0, len(v学生们))
	for _, v := range v学生们 {
		a学生列表 = append(a学生列表, &Req学生信息{
			ID:  int64(v.ID),
			V名字: v.V名字,
			V年龄: v.V年龄,
			V班级: v.V班级,
		})
	}
	return &Res学生列表{V学生列表: a学生列表, V总数: int32(total)}, nil
}
