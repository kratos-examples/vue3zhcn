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

type Student struct {
	ID        int64
	Name      string
	Age       int32
	ClassName string
}

type StudentUsecase struct {
	data *data.Data
	repo *gormrepo.Repo[models.T学生, *models.T学生Columns]
	log  *log.Helper
}

func NewStudentUsecase(data *data.Data, logger log.Logger) *StudentUsecase {
	return &StudentUsecase{
		data: data,
		repo: gormrepo.NewRepo(gormclass.Use(&models.T学生{})),
		log:  log.NewHelper(logger),
	}
}

func (uc *StudentUsecase) CreateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
	must.Nice(s.Name)

	db := uc.data.DB()

	var v学生 *models.T学生

	if erk, err := gormkratos.Transaction(ctx, db, func(db *gorm.DB) *errors.Error {
		v学生 = &models.T学生{
			V名字: s.Name,
			V年龄: s.Age,
			V班级: s.ClassName,
		}
		if err := uc.repo.With(ctx, db).Create(v学生); err != nil {
			return errors.New(500, "DB_ERROR", err.Error())
		}
		return nil
	}); err != nil {
		if erk != nil {
			return nil, ebzkratos.New(erk)
		}
		return nil, ebzkratos.New(pb.ErrorServerError("tx: %v", err))
	}
	return &Student{
		ID:        int64(v学生.ID),
		Name:      v学生.V名字,
		Age:       v学生.V年龄,
		ClassName: v学生.V班级,
	}, nil
}

func (uc *StudentUsecase) UpdateStudent(ctx context.Context, s *Student) (*Student, *ebzkratos.Ebz) {
	must.True(s.ID > 0)
	must.Nice(s.Name)

	db := uc.data.DB()

	if err := uc.repo.With(ctx, db).UpdatesM(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(uint(s.ID)))
	}, func(cls *models.T学生Columns) gormcnm.ColumnValueMap {
		return cls.Kw(cls.V名字.Kv(s.Name)).Kw(cls.V年龄.Kv(s.Age)).Kw(cls.V班级.Kv(s.ClassName))
	}); err != nil {
		return nil, ebzkratos.New(pb.ErrorServerError("update: %v", err))
	}

	return s, nil
}

func (uc *StudentUsecase) DeleteStudent(ctx context.Context, id int64) *ebzkratos.Ebz {
	must.True(id > 0)

	db := uc.data.DB()

	if err := uc.repo.With(ctx, db).DeleteW(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(uint(id)))
	}); err != nil {
		return ebzkratos.New(pb.ErrorServerError("delete: %v", err))
	}
	return nil
}

func (uc *StudentUsecase) GetStudent(ctx context.Context, id int64) (*Student, *ebzkratos.Ebz) {
	must.True(id > 0)

	db := uc.data.DB()

	v学生, erb := uc.repo.With(ctx, db).FirstE(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Where(cls.ID.Eq(uint(id)))
	})
	if erb != nil {
		if erb.NotExist {
			return nil, ebzkratos.New(pb.ErrorServerError("not found: %v", erb.Cause))
		}
		return nil, ebzkratos.New(pb.ErrorServerError("db: %v", erb.Cause))
	}

	return &Student{
		ID:        int64(v学生.ID),
		Name:      v学生.V名字,
		Age:       v学生.V年龄,
		ClassName: v学生.V班级,
	}, nil
}

func (uc *StudentUsecase) ListStudents(ctx context.Context, page int32, pageSize int32) ([]*Student, int32, *ebzkratos.Ebz) {
	db := uc.data.DB()

	v学生们, err := uc.repo.With(ctx, db).Find(func(db *gorm.DB, cls *models.T学生Columns) *gorm.DB {
		return db.Order(cls.ID.Ob("DESC").Ox())
	})
	if err != nil {
		return nil, 0, ebzkratos.New(pb.ErrorServerError("list: %v", err))
	}

	items := make([]*Student, 0, len(v学生们))
	for _, v := range v学生们 {
		items = append(items, &Student{
			ID:        int64(v.ID),
			Name:      v.V名字,
			Age:       v.V年龄,
			ClassName: v.V班级,
		})
	}
	return items, int32(len(items)), nil
}
