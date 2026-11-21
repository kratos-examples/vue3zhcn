package models

type T学生 struct {
	ID  int64  `gorm:"primaryKey;autoIncrement"`
	V名字 string `gorm:"column:name;size:128;not null"`
	V年龄 int32  `gorm:"column:age"`
	V班级 string `gorm:"column:class_name;size:128"`
}

func (*T学生) TableName() string {
	return "students"
}
