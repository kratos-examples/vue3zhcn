package service

import (
	"context"

	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
	"github.com/yylego/kratos-examples/demo2kratos/internal/biz"
)

type ArticleService struct {
	pb.UnimplementedArticleServiceServer

	uc *biz.Uc文章管理
}

func NewArticleService(uc *biz.Uc文章管理) *ArticleService {
	return &ArticleService{uc: uc}
}

func (s *ArticleService) CreateArticle(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleReply, error) {
	if req.Title == "" {
		return nil, pb.ErrorBadParam("TITLE IS REQUIRED")
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

func (s *ArticleService) UpdateArticle(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	if req.Title == "" {
		return nil, pb.ErrorBadParam("TITLE IS REQUIRED")
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

func (s *ArticleService) DeleteArticle(ctx context.Context, req *pb.DeleteArticleRequest) (*pb.DeleteArticleReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	if ebz := s.uc.Xqt删除文章(ctx, req.Id); ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.DeleteArticleReply{Success: true}, nil
}

func (s *ArticleService) GetArticle(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorBadParam("ID IS REQUIRED")
	}
	v, ebz := s.uc.Get获取文章(ctx, req.Id)
	if ebz != nil {
		return nil, ebz.Erk
	}
	return &pb.GetArticleReply{Article: &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号}}, nil
}

func (s *ArticleService) ListArticles(ctx context.Context, req *pb.ListArticlesRequest) (*pb.ListArticlesReply, error) {
	if req.Page < 1 {
		return nil, pb.ErrorBadParam("PAGE MUST BE POSITIVE")
	}
	if req.PageSize < 1 {
		return nil, pb.ErrorBadParam("PAGE_SIZE MUST BE POSITIVE")
	}
	a文章列表, count, ebz := s.uc.Get文章列表(ctx, req.Page, req.PageSize)
	if ebz != nil {
		return nil, ebz.Erk
	}
	items := make([]*pb.ArticleInfo, 0, len(a文章列表))
	for _, v := range a文章列表 {
		items = append(items, &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号})
	}
	return &pb.ListArticlesReply{Articles: items, Count: count}, nil
}

func (s *ArticleService) ListStudentArticles(ctx context.Context, req *pb.ListStudentArticlesRequest) (*pb.ListArticlesReply, error) {
	if req.StudentId <= 0 {
		return nil, pb.ErrorBadParam("STUDENT_ID IS REQUIRED")
	}
	if req.Page < 1 {
		return nil, pb.ErrorBadParam("PAGE MUST BE POSITIVE")
	}
	if req.PageSize < 1 {
		return nil, pb.ErrorBadParam("PAGE_SIZE MUST BE POSITIVE")
	}
	a文章列表, count, ebz := s.uc.Get学生文章列表(ctx, req.StudentId, req.Page, req.PageSize)
	if ebz != nil {
		return nil, ebz.Erk
	}
	items := make([]*pb.ArticleInfo, 0, len(a文章列表))
	for _, v := range a文章列表 {
		items = append(items, &pb.ArticleInfo{Id: v.ID, Title: v.V标题, Content: v.V内容, StudentId: v.V学生编号})
	}
	return &pb.ListArticlesReply{Articles: items, Count: count}, nil
}
