package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
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
)

type Req学生信息 struct {
	ID int64
	V名字 string
	V年龄 int32
	V班级 string
}

type Uc学生管理 struct {
	data *data.Data
	repo学生 *gormrepo.Repo[models.T学生, *models.T学生Columns]
	log  *log.Helper
}

func NewUc学生管理(data *data.Data, logger log.Logger) *Uc学生管理 {
	return &Uc学生管理{
		data: data,
		repo学生: gormrepo.NewRepo(gormclass.Use(&models.T学生{})),
		log:  log.NewHelper(logger),
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
			return errors.New(500, "DB_ERROR", err.Error())
		}
		return nil
	}); err != nil {
		if erk != nil {
			return nil, ebzkratos.New(erk)
		}
		return nil, ebzkratos.New(pb.ErrorServerError("tx: %v", err))
	}
	return &Req学生信息{
		ID: int64(v学生.ID),
		V名字: v学生.V名字,
		V年龄: v学生.V年龄,
		V班级: v学生.V班级,
	}, nil
}

func (uc *Uc学生管理) Xqt更新学生(ctx context.Context, req *Req学生信息) (*Req学生信息, *ebzkratos.Ebz) {
	must.True(req.ID > 0)
	must.Nice(req.V名字)

	db := uc.data.DB()

	if err := uc.repo学生.With(ctx, db).UpdatesM(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(uint(req.ID)))
	}, func(cls *models.T学生Columns) gormcnm.ColumnValueMap {
		return cls.Kw(cls.V名字.Kv(req.V名字)).Kw(cls.V年龄.Kv(req.V年龄)).Kw(cls.V班级.Kv(req.V班级))
	}); err != nil {
		return nil, ebzkratos.New(pb.ErrorServerError("update: %v", err))
	}

	return req, nil
}

func (uc *Uc学生管理) Xqt删除学生(ctx context.Context, id int64) *ebzkratos.Ebz {
	must.True(id > 0)

	db := uc.data.DB()

	if err := uc.repo学生.With(ctx, db).DeleteW(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(uint(id)))
	}); err != nil {
		return ebzkratos.New(pb.ErrorServerError("delete: %v", err))
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
			return nil, ebzkratos.New(pb.ErrorServerError("not found: %v", erb.Cause))
		}
		return nil, ebzkratos.New(pb.ErrorServerError("db: %v", erb.Cause))
	}

	return &Req学生信息{
		ID: int64(v学生.ID),
		V名字: v学生.V名字,
		V年龄: v学生.V年龄,
		V班级: v学生.V班级,
	}, nil
}

func (uc *Uc学生管理) Get学生列表(ctx context.Context, page int32, pageSize int32) ([]*Req学生信息, int32, *ebzkratos.Ebz) {
	db := uc.data.DB()

	v学生们, err := uc.repo学生.With(ctx, db).Find(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Order(cls.ID.Ob("DESC").Ox())
	})
	if err != nil {
		return nil, 0, ebzkratos.New(pb.ErrorServerError("list: %v", err))
	}

	a学生列表 := make([]*Req学生信息, 0, len(v学生们))
	for _, v := range v学生们 {
		a学生列表 = append(a学生列表, &Req学生信息{
			ID: int64(v.ID),
			V名字: v.V名字,
			V年龄: v.V年龄,
			V班级: v.V班级,
		})
	}
	return a学生列表, int32(len(a学生列表)), nil
}
