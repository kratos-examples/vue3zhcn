package service

import (
	"context"

	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
	"github.com/yylego/kratos-examples/demo2kratos/internal/biz"
)

type Svc文章服务 struct {
	pb.UnimplementedArticleServiceServer

	uc *biz.Uc文章管理
}

func NewSvc文章服务(uc *biz.Uc文章管理) *Svc文章服务 {
	return &Svc文章服务{uc: uc}
}

func (s *Svc文章服务) CreateArticle(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleReply, error) {
	if req.Title == "" {
		return nil, pb.ErrorBadParam("TITLE IS REQUIRED")
	}
	if req.StudentId <= 0 {
		return nil, pb.ErrorBadParam("STUDENT_ID IS REQUIRED")
	}
	v, ebz := s.uc.Xqt创建文章(ctx, &biz.Req文章信息{
		V标题:   req.Title,
		V内容:   req.Content,
		V学生编号: req.StudentId,
	})
	if ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.CreateArticleReply{Article: &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号}}, nil
}

func (s *Svc文章服务) UpdateArticle(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	if req.Title == "" {
		return nil, pb.ErrorBadParam("TITLE IS REQUIRED")
	}
	if req.StudentId <= 0 {
		return nil, pb.ErrorBadParam("STUDENT_ID IS REQUIRED")
	}
	v, ebz := s.uc.Xqt更新文章(ctx, &biz.Req文章信息{
		ID:    req.Id,
		V标题:   req.Title,
		V内容:   req.Content,
		V学生编号: req.StudentId,
	})
	if ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.UpdateArticleReply{Article: &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号}}, nil
}

func (s *Svc文章服务) DeleteArticle(ctx context.Context, req *pb.DeleteArticleRequest) (*pb.DeleteArticleReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	if ebz := s.uc.Xqt删除文章(ctx, req.Id); ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.DeleteArticleReply{Success: true}, nil
}

func (s *Svc文章服务) GetArticle(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	v, ebz := s.uc.Get获取文章(ctx, req.Id)
	if ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.GetArticleReply{Article: &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号}}, nil
}

func (s *Svc文章服务) ListArticles(ctx context.Context, req *pb.ListArticlesRequest) (*pb.ListArticlesReply, error) {
	if req.Page < 1 {
		return nil, pb.ErrorBadParam("PAGE MUST BE POSITIVE")
	}
	if req.PageSize < 1 {
		return nil, pb.ErrorBadParam("PAGE_SIZE MUST BE POSITIVE")
	}
	res, ebz := s.uc.Get文章列表(ctx, &biz.Req文章列表{V页码: req.Page, V每页大小: req.PageSize})
	if ebz != nil {
		return nil, ebz.Erk
	}
	items := make([]*pb.ArticleInfo, 0, len(res.V文章列表))
	for _, v := range res.V文章列表 {
		items = append(items, &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号})
	}
	return &pb.ListArticlesReply{Articles: items, Count: res.V总数}, nil
}

func (s *Svc文章服务) ListStudentArticles(ctx context.Context, req *pb.ListStudentArticlesRequest) (*pb.ListArticlesReply, error) {
	if req.StudentId <= 0 {
		return nil, pb.ErrorBadParam("STUDENT_ID IS REQUIRED")
	}
	if req.Page < 1 {
		return nil, pb.ErrorBadParam("PAGE MUST BE POSITIVE")
	}
	if req.PageSize < 1 {
		return nil, pb.ErrorBadParam("PAGE_SIZE MUST BE POSITIVE")
	}
	res, ebz := s.uc.Get学生文章列表(ctx, &biz.Req学生文章列表{V学生编号: req.StudentId, V页码: req.Page, V每页大小: req.PageSize})
	if ebz != nil {
		return nil, ebz.Erk
	}
	items := make([]*pb.ArticleInfo, 0, len(res.V文章列表))
	for _, v := range res.V文章列表 {
		items = append(items, &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号})
	}
	return &pb.ListArticlesReply{Articles: items, Count: res.V总数}, nil
}
