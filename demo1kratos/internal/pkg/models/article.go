package models

type T文章 struct {
	ID    int64  `gorm:"primaryKey;autoIncrement"`
	V标题   string `gorm:"column:title;size:256;not null"`
	V内容   string `gorm:"column:content;type:text"`
	V学生编号 int64  `gorm:"column:student_id;index:idx_articles_student_id"`
}

func (*T文章) TableName() string {
	return "articles"
}
