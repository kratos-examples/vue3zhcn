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
	log    *slog.Logger
}

func NewUc学生管理(data *data.Data, logger *slog.Logger) *Uc学生管理 {
	return &Uc学生管理{
		data:   data,
		repo学生: gormrepo.NewRepo(gormclass.Use(&models.T学生{})),
		repo文章: gormrepo.NewRepo(gormclass.Use(&models.T文章{})),
		log:    logger,
	}
}

func (uc *Uc学生管理) Xqt创建学生(ctx context.Context, req *Req学生信息) (*Req学生信息, *ebzkratos.Ebz) {
	must.Nice(req.V名字)

	db := uc.data.DB()

	var v学生 *models.T学生

	if erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		v学生 = &models.T学生{
			V名字: req.V名字,
			V年龄: req.V年龄,
			V班级: req.V班级,
		}
		if err := uc.repo学生.With(ctx, db).Create(v学生); err != nil {
			return pb.ErrorStudentCreateFailure("create student: %v", err)
		}
		return nil
	}); err != nil {
		if erk != nil {
			return nil, ebzkratos.New(erk)
		}
		return nil, ebzkratos.New(pb.ErrorTxError("tx: %v", err))
	}
	return &Req学生信息{
		ID:  int64(v学生.ID),
		V名字: v学生.V名字,
		V年龄: v学生.V年龄,
		V班级: v学生.V班级,
	}, nil
}

func (uc *Uc学生管理) Xqt更新学生(ctx context.Context, req *Req学生信息) (*Req学生信息, *ebzkratos.Ebz) {
	must.True(req.ID > 0)
	must.Nice(req.V名字)

	db := uc.data.DB()

	// 先确认学生存在，对齐桩子：查不到返回 StudentNotFound 而非静默成功
	if _, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(uint(req.ID)))
	}); erb != nil {
		if erb.NotExist {
			return nil, ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", req.ID))
		}
		return nil, ebzkratos.New(pb.ErrorDbError("get student: %v", erb.Cause))
	}

	if err := uc.repo学生.With(ctx, db).UpdatesM(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(uint(req.ID)))
	}, func(cls *models.T学生Columns) gormcnm.ColumnValueMap {
		return cls.Kw(cls.V名字.Kv(req.V名字)).Kw(cls.V年龄.Kv(req.V年龄)).Kw(cls.V班级.Kv(req.V班级))
	}); err != nil {
		return nil, ebzkratos.New(pb.ErrorDbError("update student: %v", err))
	}

	return req, nil
}

func (uc *Uc学生管理) Xqt删除学生(ctx context.Context, id int64) *ebzkratos.Ebz {
	must.True(id > 0)

	db := uc.data.DB()

	// 翻译桩子的原子级联删除，全在一个事务里：
	//   ① FOR UPDATE 锁住学生行，删除期间挡住并发新建文章（对方持互斥的 FOR SHARE 锁）；
	//   ② 先删该学生名下的文章（子表在前）；
	//   ③ 再删学生本身（父表在后）。
	var b未找到 bool
	if erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		if _, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
			return db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).Where(cls.ID.Eq(uint(id)))
		}); erb != nil {
			if erb.NotExist {
				b未找到 = true
				return nil
			}
			return pb.ErrorDbError("get student: %v", erb.Cause)
		}
		if err := uc.repo文章.With(ctx, db).DeleteW(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
			return db.Where(cls.V学生编号.Eq(id))
		}); err != nil {
			return pb.ErrorDbError("delete articles: %v", err)
		}
		if err := uc.repo学生.With(ctx, db).DeleteW(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
			return db.Where(cls.ID.Eq(uint(id)))
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
		return ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", id))
	}
	return nil
}

func (uc *Uc学生管理) Get获取学生(ctx context.Context, id int64) (*Req学生信息, *ebzkratos.Ebz) {
	must.True(id > 0)

	db := uc.data.DB()

	v学生, erb := uc.repo学生.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(uint(id)))
	})
	if erb != nil {
		if erb.NotExist {
			return nil, ebzkratos.New(pb.ErrorStudentNotFound("student %d not found", id))
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

func (uc *Uc学生管理) Get学生列表(ctx context.Context, page int32, pageSize int32) ([]*Req学生信息, int32, *ebzkratos.Ebz) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

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
			Offset: int((page - 1) * pageSize),
			Limit:  int(pageSize),
		},
	)
	if err != nil {
		return nil, 0, ebzkratos.New(pb.ErrorDbError("list students: %v", err))
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
	return a学生列表, int32(total), nil
}
