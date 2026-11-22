package service

import (
	"context"
	"errors"
	"io"

	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
	"github.com/yylego/kratos-examples/demo1kratos/internal/biz"
)

type Svc学生服务 struct {
	pb.UnimplementedStudentServiceServer

	uc *biz.Uc学生管理
}

func NewSvc学生服务(uc *biz.Uc学生管理) *Svc学生服务 {
	return &Svc学生服务{uc: uc}
}

func (s *Svc学生服务) CreateStudent(ctx context.Context, req *pb.CreateStudentRequest) (*pb.CreateStudentReply, error) {
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

func (s *Svc学生服务) UpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) (*pb.UpdateStudentReply, error) {
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

func (s *Svc学生服务) DeleteStudent(ctx context.Context, req *pb.DeleteStudentRequest) (*pb.DeleteStudentReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	if ebz := s.uc.Xqt删除学生(ctx, req.Id); ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.DeleteStudentReply{Success: true}, nil
}

func (s *Svc学生服务) GetStudent(ctx context.Context, req *pb.GetStudentRequest) (*pb.GetStudentReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	v, ebz := s.uc.Get获取学生(ctx, req.Id)
	if ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.GetStudentReply{Student: &pb.StudentInfo{Id: v.ID, Name: v.V名字, Age: v.V年龄, ClassName: v.V班级}}, nil
}

func (s *Svc学生服务) ListStudents(ctx context.Context, req *pb.ListStudentsRequest) (*pb.ListStudentsReply, error) {
	if req.Page < 1 {
		return nil, pb.ErrorBadParam("PAGE MUST BE POSITIVE")
	}
	if req.PageSize < 1 {
		return nil, pb.ErrorBadParam("PAGE_SIZE MUST BE POSITIVE")
	}
	res, ebz := s.uc.Get学生列表(ctx, &biz.Req学生列表{V页码: req.Page, V每页大小: req.PageSize})
	if ebz != nil {
		return nil, ebz.Erk
	}
	items := make([]*pb.StudentInfo, 0, len(res.V学生列表))
	for _, v := range res.V学生列表 {
		items = append(items, &pb.StudentInfo{Id: v.ID, Name: v.V名字, Age: v.V年龄, ClassName: v.V班级})
	}
	return &pb.ListStudentsReply{Students: items, Count: res.V总数}, nil
}

// LookupStudents answers each student id that comes in with that student, until the client stops sending.
// A failed lookup (such as a missing student) ends the stream with that error.
//
// LookupStudents 对每个发来的学生编号回复那个学生，直到客户端不再发送；某次查询失败（比如学生不存在）就带着那个错误结束这条流。
func (s *Svc学生服务) LookupStudents(stream pb.StudentService_LookupStudentsServer) error {
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		reply, err := s.GetStudent(stream.Context(), req)
		if err != nil {
			return err
		}
		if err := stream.Send(reply); err != nil {
			return err
		}
	}
}
