package models

import "gorm.io/gorm"

type T学生 struct {
	gorm.Model
	V名字 string `gorm:"column:name;type:varchar(255)" cnm:"V名字"`
	V年龄 int32  `gorm:"column:age;type:int" cnm:"V年龄"`
	V班级 string `gorm:"column:class_name;type:varchar(255)" cnm:"V班级"`
}

func (*T学生) TableName() string {
	return "students"
}
