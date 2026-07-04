package service

import (
	"context"

	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
	"github.com/yylego/kratos-examples/demo1kratos/internal/biz"
)

type StudentService struct {
	pb.UnimplementedStudentServiceServer

	uc *biz.Uc学生管理
}

func NewStudentService(uc *biz.Uc学生管理) *StudentService {
	return &StudentService{uc: uc}
}

func (s *StudentService) CreateStudent(ctx context.Context, req *pb.CreateStudentRequest) (*pb.CreateStudentReply, error) {
	if req.Name == "" {
		return nil, pb.ErrorBadParam("NAME IS REQUIRED")
	}
	v, ebz := s.uc.Xqt创建学生(ctx, &biz.Req学生信息{
		V名字: req.Name,
		V年龄: req.Age,
		V班级: req.ClassName,
	})
	if ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.CreateStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.V名字, Age: v.V年龄, ClassName: v.V班级}}, nil
}

func (s *StudentService) UpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) (*pb.UpdateStudentReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	if req.Name == "" {
		return nil, pb.ErrorBadParam("NAME IS REQUIRED")
	}
	v, ebz := s.uc.Xqt更新学生(ctx, &biz.Req学生信息{
		ID:  req.Id,
		V名字: req.Name,
		V年龄: req.Age,
		V班级: req.ClassName,
	})
	if ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.UpdateStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.V名字, Age: v.V年龄, ClassName: v.V班级}}, nil
}

func (s *StudentService) DeleteStudent(ctx context.Context, req *pb.DeleteStudentRequest) (*pb.DeleteStudentReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	if ebz := s.uc.Xqt删除学生(ctx, req.Id); ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.DeleteStudentReply{Success: true}, nil
}

func (s *StudentService) GetStudent(ctx context.Context, req *pb.GetStudentRequest) (*pb.GetStudentReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	v, ebz := s.uc.Get获取学生(ctx, req.Id)
	if ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.GetStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.V名字, Age: v.V年龄, ClassName: v.V班级}}, nil
}

func (s *StudentService) ListStudents(ctx context.Context, req *pb.ListStudentsRequest) (*pb.ListStudentsReply, error) {
	a学生列表, count, ebz := s.uc.Get学生列表(ctx, req.Page, req.PageSize)
	if ebz != nil {
		return nil, ebz.Erk
	}
	items := make([]*pb.StudentInfo, 0, len(a学生列表))
	for _, v := range a学生列表 {
		items = append(items, &pb.StudentInfo{Id: v.ID, Name: v.V名字, Age: v.V年龄, ClassName: v.V班级})
	}
	return &pb.ListStudentsReply{Students: items, Count: count}, nil
}
