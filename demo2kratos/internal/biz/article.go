package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/yylego/gormcnm"
	"github.com/yylego/gormrepo"
	"github.com/yylego/gormrepo/gormclass"
	"github.com/yylego/kratos-ebz/ebzkratos"
	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
	"github.com/yylego/kratos-examples/demo2kratos/internal/data"
	"github.com/yylego/kratos-examples/demo2kratos/internal/pkg/models"
	"github.com/yylego/kratos-gorm/gormkratos"
	"github.com/yylego/must"
	"gorm.io/gorm"
)

type Req文章信息 struct {
	ID   int64
	V标题   string
	V内容   string
	V学生编号 int64
}

type Uc文章管理 struct {
	data *data.Data
	repo文章 *gormrepo.Repo[models.T文章, *models.T文章Columns]
	log  *log.Helper
}

func NewUc文章管理(data *data.Data, logger log.Logger) *Uc文章管理 {
	return &Uc文章管理{
		data: data,
		repo文章: gormrepo.NewRepo(gormclass.Use(&models.T文章{})),
		log:  log.NewHelper(logger),
	}
}

func (uc *Uc文章管理) Xqt创建文章(ctx context.Context, req *Req文章信息) (*Req文章信息, *ebzkratos.Ebz) {
	must.Nice(req.V标题)

	db := uc.data.DB()

	var v文章 *models.T文章

	if erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		v文章 = &models.T文章{
			V标题:   req.V标题,
			V内容:   req.V内容,
			V学生编号: req.V学生编号,
		}
		if err := uc.repo文章.With(ctx, db).Create(v文章); err != nil {
			return errors.New(500, "DB_ERROR", err.Error())
		}
		return nil
	}); err != nil {
		if erk != nil {
			return nil, ebzkratos.New(erk)
		}
		return nil, ebzkratos.New(pb.ErrorServerError("tx: %v", err))
	}
	return &Req文章信息{
		ID:   int64(v文章.ID),
		V标题:   v文章.V标题,
		V内容:   v文章.V内容,
		V学生编号: v文章.V学生编号,
	}, nil
}

func (uc *Uc文章管理) Xqt更新文章(ctx context.Context, req *Req文章信息) (*Req文章信息, *ebzkratos.Ebz) {
	must.True(req.ID > 0)
	must.Nice(req.V标题)

	db := uc.data.DB()

	if err := uc.repo文章.With(ctx, db).UpdatesM(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(uint(req.ID)))
	}, func(cls *models.T文章Columns) gormcnm.ColumnValueMap {
		return cls.Kw(cls.V标题.Kv(req.V标题)).Kw(cls.V内容.Kv(req.V内容)).Kw(cls.V学生编号.Kv(req.V学生编号))
	}); err != nil {
		return nil, ebzkratos.New(pb.ErrorServerError("update: %v", err))
	}

	return req, nil
}

func (uc *Uc文章管理) Xqt删除文章(ctx context.Context, id int64) *ebzkratos.Ebz {
	must.True(id > 0)

	db := uc.data.DB()

	if err := uc.repo文章.With(ctx, db).DeleteW(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(uint(id)))
	}); err != nil {
		return ebzkratos.New(pb.ErrorServerError("delete: %v", err))
	}
	return nil
}

func (uc *Uc文章管理) Get获取文章(ctx context.Context, id int64) (*Req文章信息, *ebzkratos.Ebz) {
	must.True(id > 0)

	db := uc.data.DB()

	v文章, erb := uc.repo文章.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(uint(id)))
	})
	if erb != nil {
		if erb.NotExist {
			return nil, ebzkratos.New(pb.ErrorServerError("not found: %v", erb.Cause))
		}
		return nil, ebzkratos.New(pb.ErrorServerError("db: %v", erb.Cause))
	}

	return &Req文章信息{
		ID:   int64(v文章.ID),
		V标题:   v文章.V标题,
		V内容:   v文章.V内容,
		V学生编号: v文章.V学生编号,
	}, nil
}

func (uc *Uc文章管理) Get文章列表(ctx context.Context, page int32, pageSize int32) ([]*Req文章信息, int32, *ebzkratos.Ebz) {
	db := uc.data.DB()

	v文章们, err := uc.repo文章.With(ctx, db).Find(func(db *gorm.DB, cls *models.T文章Columns) *gorm.DB {
		return db.Order(cls.ID.Ob("DESC").Ox())
	})
	if err != nil {
		return nil, 0, ebzkratos.New(pb.ErrorServerError("list: %v", err))
	}

	a文章列表 := make([]*Req文章信息, 0, len(v文章们))
	for _, v := range v文章们 {
		a文章列表 = append(a文章列表, &Req文章信息{
			ID:   int64(v.ID),
			V标题:   v.V标题,
			V内容:   v.V内容,
			V学生编号: v.V学生编号,
		})
	}
	return a文章列表, int32(len(a文章列表)), nil
}
