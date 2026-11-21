// Code generated using gormcngen. DO NOT EDIT.
// This file was auto generated via github.com/yylego/gormcngen

//go:build !gormcngen_generate

// Generated from: gormcnm.gen_test.go:32 -> models_test.TestGenerateColumns
// ========== GORMCNGEN:DO-NOT-EDIT-MARKER:END ==========

// Code generated using gormcngen. DO NOT EDIT.
// This file was auto generated via github.com/yylego/gormcngen

package models

import (
	"github.com/yylego/gormcnm"
)

func (c *T文章) Columns() *T文章Columns {
	return c.TableColumns(gormcnm.NewPlainDecoration())
}

type T文章Columns struct {
	// Auto-generated: embedding operation functions to make it simple to use. DO NOT EDIT. // 自动生成：嵌入操作函数便于使用。请勿编辑。
	gormcnm.ColumnOperationClass
	// Auto-generated: column names and types in database table. DO NOT EDIT. // 自动生成：数据库表的列名和类型。请勿编辑。
	ID    gormcnm.ColumnName[int64]
	V标题   gormcnm.ColumnName[string]
	V内容   gormcnm.ColumnName[string]
	V学生编号 gormcnm.ColumnName[int64]
}

func (c *T学生) Columns() *T学生Columns {
	return c.TableColumns(gormcnm.NewPlainDecoration())
}

type T学生Columns struct {
	// Auto-generated: embedding operation functions to make it simple to use. DO NOT EDIT. // 自动生成：嵌入操作函数便于使用。请勿编辑。
	gormcnm.ColumnOperationClass
	// Auto-generated: column names and types in database table. DO NOT EDIT. // 自动生成：数据库表的列名和类型。请勿编辑。
	ID  gormcnm.ColumnName[int64]
	V名字 gormcnm.ColumnName[string]
	V年龄 gormcnm.ColumnName[int32]
	V班级 gormcnm.ColumnName[string]
}

func (c *T文章) TableColumns(decoration gormcnm.ColumnNameDecoration) *T文章Columns {
	return &T文章Columns{
		// Auto-generated: column mapping in table operations. DO NOT EDIT. // 自动生成：表操作的列映射。请勿编辑。
		ID:    gormcnm.Cmn(c.ID, "id", decoration),
		V标题:   gormcnm.Cmn(c.V标题, "title", decoration),
		V内容:   gormcnm.Cmn(c.V内容, "content", decoration),
		V学生编号: gormcnm.Cmn(c.V学生编号, "student_id", decoration),
	}
}

func (c *T学生) TableColumns(decoration gormcnm.ColumnNameDecoration) *T学生Columns {
	return &T学生Columns{
		// Auto-generated: column mapping in table operations. DO NOT EDIT. // 自动生成：表操作的列映射。请勿编辑。
		ID:  gormcnm.Cmn(c.ID, "id", decoration),
		V名字: gormcnm.Cmn(c.V名字, "name", decoration),
		V年龄: gormcnm.Cmn(c.V年龄, "age", decoration),
		V班级: gormcnm.Cmn(c.V班级, "class_name", decoration),
	}
}
