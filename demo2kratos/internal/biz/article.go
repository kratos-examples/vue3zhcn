package biz

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/yylego/gormcnm"
	"github.com/yylego/gormrepo"
	"github.com/yylego/gormrepo/gormclass"
	"github.com/yylego/kratos-ebz/ebzkratos"
	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
	"github.com/yylego/kratos-examples/demo2kratos/internal/pkg/models"
	"github.com/yylego/kratos-gorm/gormkratos"
	"github.com/yylego/must"
	"github.com/yylego/rese"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Req文章信息 struct {
	ID    int64
	V标题   string
	V内容   string
	V学生编号 int64
}

type Uc文章管理 struct {
	data   *data.Data
	repo文章 *gormrepo.Repo[models.T文章, *models.T文章Columns]
	repo学生 *gormrepo.Repo[models.T学生, *models.T学生Columns]
	slog   *slog.Logger
}

func NewUc文章管理(data *data.Data, logger *slog.Logger) (*Uc文章管理, error) {
	// 与学生服务共用一个库：在这里把两张表都建好。
	if err := data.DB().AutoMigrate(&models.T文章{}, &models.T学生{}); err != nil {
		return nil, err
	}
	return &Uc文章管理{
		data:   data,
		repo文章: gormrepo.NewRepo(gormclass.Use(&models.T文章{})),
		repo学生: gormrepo.NewRepo(gormclass.Use(&models.T学生{})),
		slog:   logger,
	}, nil
}

func (uc *Uc文章管理) Xqt创建文章(ctx context.Context, req *Req文章信息) (*Req文章信息, *ebzkratos.Ebz) {
	must.Nice(req.V标题)
	must.True(req.V学生编号 > 0)

	db := uc.data.DB()

	var v文章 *models.T文章

	// 翻译桩子：在一个事务里 FOR SHARE 锁住学生行再插文章，挡住并发的
	// DeleteStudent（它持 FOR UPDATE），绝不创建指向"正被删除的学生"的文章。
	if erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		if _, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
			return db.Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).Where(cls.ID.Eq(req.V学生编号))
		}); erb != nil {
			if erb.NotExist {
				return pb.ErrorBadParam("student %d does not exist", req.V学生编号)
			}
			return pb.ErrorDbError("get student: %v", erb.Cause)
		}
		v文章 = &models.T文章{
			V标题:   req.V标题,
			V内容:   req.V内容,
			V学生编号: req.V学生编号,
		}
		if err := uc.repo文章.With(ctx, db).Create(v文章); err != nil {
			return pb.ErrorArticleCreateFailure("create article: %v", err)
		}
		return nil
	}); err != nil {
		if erk != nil {
			return nil, ebzkratos.New(erk)
		}
		return nil, ebzkratos.New(pb.ErrorTxError("tx: %v", err))
	}
	res := &Req文章信息{
		ID:    int64(v文章.ID),
		V标题:   v文章.V标题,
		V内容:   v文章.V内容,
		V学生编号: v文章.V学生编号,
	}
	uc.slog.InfoContext(ctx, "created article", "id", res.ID, "student_id", res.V学生编号)
	return res, nil
}

func (uc *Uc文章管理) Xqt更新文章(ctx context.Context, req *Req文章信息) (*Req文章信息, *ebzkratos.Ebz) {
	must.True(req.ID > 0)
	must.Nice(req.V标题)
	must.True(req.V学生编号 > 0)

	db := uc.data.DB()

	var v文章 *models.T文章
	// FOR SHARE 锁住新归属学生（对齐桩子），再更新文章、靠 RowsAffected 判文章不存在，最后 reload——忠实对齐桩子。
	if erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		if _, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
			return db.Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).Where(cls.ID.Eq(req.V学生编号))
		}); erb != nil {
			if erb.NotExist {
				return pb.ErrorBadParam("student %d does not exist", req.V学生编号)
			}
			return pb.ErrorDbError("get student: %v", erb.Cause)
		}
		upd := uc.repo文章.Wrap(ctx, db).UpdatesM(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
			return db.Where(cls.ID.Eq(req.ID))
		}, func(cls *models.T文章Columns) gormcnm.ColumnValueMap {
			return cls.Kw(cls.V标题.Kv(req.V标题)).Kw(cls.V内容.Kv(req.V内容)).Kw(cls.V学生编号.Kv(req.V学生编号))
		})
		if upd.Error != nil {
			return pb.ErrorDbError("update article: %v", upd.Error)
		}
		if upd.RowsAffected == 0 {
			return pb.ErrorArticleNotFound("article %d not found", req.ID)
		}
		got, erb := uc.repo文章.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
			return db.Where(cls.ID.Eq(req.ID))
		})
		if erb != nil {
			return pb.ErrorDbError("reload article: %v", erb.Cause)
		}
		v文章 = got
		return nil
	}); err != nil {
		if erk != nil {
			return nil, ebzkratos.New(erk)
		}
		return nil, ebzkratos.New(pb.ErrorTxError("tx: %v", err))
	}

	return &Req文章信息{
		ID:    int64(v文章.ID),
		V标题:   v文章.V标题,
		V内容:   v文章.V内容,
		V学生编号: v文章.V学生编号,
	}, nil
}

func (uc *Uc文章管理) Xqt删除文章(ctx context.Context, n编号 int64) *ebzkratos.Ebz {
	must.True(n编号 > 0)

	db := uc.data.DB()

	// 忠实翻译桩子：单条删除后靠 RowsAffected 判不存在（0 行→ArticleNotFound）。
	del := uc.repo文章.Wrap(ctx, db).DeleteW(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(n编号))
	})
	if del.Error != nil {
		return ebzkratos.New(pb.ErrorDbError("delete article: %v", del.Error))
	}
	if del.RowsAffected == 0 {
		return ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", n编号))
	}
	uc.slog.InfoContext(ctx, "deleted article", "id", n编号)
	return nil
}

func (uc *Uc文章管理) Get获取文章(ctx context.Context, n编号 int64) (*Req文章信息, *ebzkratos.Ebz) {
	must.True(n编号 > 0)

	db := uc.data.DB()

	v文章, erb := uc.repo文章.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(n编号))
	})
	if erb != nil {
		if erb.NotExist {
			return nil, ebzkratos.New(pb.ErrorArticleNotFound("article %d not found", n编号))
		}
		return nil, ebzkratos.New(pb.ErrorDbError("get article: %v", erb.Cause))
	}

	return &Req文章信息{
		ID:    int64(v文章.ID),
		V标题:   v文章.V标题,
		V内容:   v文章.V内容,
		V学生编号: v文章.V学生编号,
	}, nil
}

// Req文章列表 / Res文章列表 分别包裹「查询文章列表」的入参（分页）与返回（当页数据+总数）。
type Req文章列表 struct {
	V页码   int32
	V每页大小 int32
}

type Res文章列表 struct {
	V文章列表 []*Req文章信息
	V总数   int32
}

func (uc *Uc文章管理) Get文章列表(ctx context.Context, req *Req文章列表) (*Res文章列表, *ebzkratos.Ebz) {
	must.True(req.V页码 >= 1)
	must.True(req.V每页大小 >= 1)

	db := uc.data.DB()

	// gormrepo FindPageAndCount 一次拿到当页数据和总行数，对齐桩子的分页+计数。
	v文章们, count, err := uc.repo文章.With(ctx, db).FindPageAndCount(
		func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
			return db
		},
		func(cls *models.T文章Columns) gormcnm.Ordering {
			return cls.ID.Ob("asc")
		},
		rese.P1(gormrepo.NewPagination(int(req.V页码), int(req.V每页大小))),
	)
	if err != nil {
		return nil, ebzkratos.New(pb.ErrorDbError("list articles: %v", err))
	}

	return &Res文章列表{V文章列表: conv文章列表(v文章们), V总数: int32(count)}, nil
}

// Req学生文章列表 包裹「查询某学生的文章列表」的入参（学生编号+分页），返回复用 Res文章列表。
type Req学生文章列表 struct {
	V学生编号 int64
	V页码   int32
	V每页大小 int32
}

// Get学生文章列表 分页返回某个学生的文章（对应 proto 的 ListStudentArticles）。
func (uc *Uc文章管理) Get学生文章列表(ctx context.Context, req *Req学生文章列表) (*Res文章列表, *ebzkratos.Ebz) {
	must.True(req.V学生编号 > 0)
	must.True(req.V页码 >= 1)
	must.True(req.V每页大小 >= 1)

	db := uc.data.DB()

	v文章们, count, err := uc.repo文章.With(ctx, db).FindPageAndCount(
		func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
			return db.Where(cls.V学生编号.Eq(req.V学生编号))
		},
		func(cls *models.T文章Columns) gormcnm.Ordering {
			return cls.ID.Ob("asc")
		},
		rese.P1(gormrepo.NewPagination(int(req.V页码), int(req.V每页大小))),
	)
	if err != nil {
		return nil, ebzkratos.New(pb.ErrorDbError("list student articles: %v", err))
	}

	return &Res文章列表{V文章列表: conv文章列表(v文章们), V总数: int32(count)}, nil
}

func conv文章列表(v文章们 []*models.T文章) []*Req文章信息 {
	a文章列表 := make([]*Req文章信息, 0, len(v文章们))
	for _, v := range v文章们 {
		a文章列表 = append(a文章列表, &Req文章信息{
			ID:    int64(v.ID),
			V标题:   v.V标题,
			V内容:   v.V内容,
			V学生编号: v.V学生编号,
		})
	}
	return a文章列表
}
