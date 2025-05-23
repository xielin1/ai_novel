package model

import (
	"context"
	"gin-template/common"
	"gorm.io/gorm/schema"
	"os"
	"reflect"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func createRootAccountIfNeed() error {
	var user User
	//if user.Status != common.UserStatusEnabled {
	if err := DB.First(&user).Error; err != nil {
		common.SysLog("no user exists, create a root user for you: username is root, password is 123456")
		hashedPassword, err := common.Password2Hash("123456")
		if err != nil {
			return err
		}
		rootUser := User{
			Username:    "root",
			Password:    hashedPassword,
			Role:        common.RoleRootUser,
			Status:      common.UserStatusEnabled,
			DisplayName: "Root User",
		}
		DB.Create(&rootUser)
	}
	return nil
}

func CountTable(tableName string) (num int64) {
	DB.Table(tableName).Count(&num)
	return
}

func InitDB() (err error) {
	var db *gorm.DB
	if os.Getenv("SQL_DSN") != "" {
		// Use MySQL
		db, err = gorm.Open(mysql.Open(os.Getenv("SQL_DSN")), &gorm.Config{
			PrepareStmt: true, // precompile SQL
		})
		db.Use(&SimpleTimePlugin{UseMilli: false})
	} else {
		// Use SQLite
		db, err = gorm.Open(sqlite.Open(common.SQLitePath), &gorm.Config{
			PrepareStmt: true, // precompile SQL
		})
		common.SysLog("SQL_DSN not set, using SQLite as database")
	}
	if err == nil {
		DB = db
		err := db.AutoMigrate(&File{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&User{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Option{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Project{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Outline{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Version{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&Referral{})
		if err != nil {
			return err
		}
		err = db.AutoMigrate(&ReferralUse{})
		if err != nil {
			return err
		}
		err = createRootAccountIfNeed()
		return err
	} else {
		common.FatalLog(err)
	}
	return err
}

func CloseDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	return err
}

type SimpleTimePlugin struct {
	UseMilli bool // 是否使用毫秒时间戳
}

func (p *SimpleTimePlugin) Name() string { return "simple_time_plugin" }

func (p *SimpleTimePlugin) Initialize(db *gorm.DB) error {
	// 注册回调
	db.Callback().Create().Before("gorm:create").Register("set_time", p.setCreateTime)
	db.Callback().Update().Before("gorm:update").Register("set_time", p.setUpdateTime)
	return nil
}

// 获取时间戳
func (p *SimpleTimePlugin) now() int64 {
	if p.UseMilli {
		return time.Now().UnixMilli()
	}
	return time.Now().Unix()
}

// 创建时设置 create_at 和 update_at
func (p *SimpleTimePlugin) setCreateTime(db *gorm.DB) {
	if db.Statement.Schema == nil {
		return
	}

	now := p.now()
	p.setField(db, "create_at", now)
	p.setField(db, "update_at", now)
}

// 更新时设置 update_at
func (p *SimpleTimePlugin) setUpdateTime(db *gorm.DB) {
	if db.Statement.Schema == nil || db.Statement.SkipHooks {
		return
	}
	p.setField(db, "update_at", p.now())
}

// 通用字段设置方法
func (p *SimpleTimePlugin) setField(db *gorm.DB, fieldName string, value int64) {
	field := db.Statement.Schema.LookUpField(fieldName)
	if field == nil {
		return
	}

	rv := reflect.ValueOf(db.Statement.Dest)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			p.setFieldValue(rv.Index(i), field, value)
		}
	case reflect.Struct:
		p.setFieldValue(rv, field, value)
	}
}

// 设置字段值
func (p *SimpleTimePlugin) setFieldValue(rv reflect.Value, field *schema.Field, value int64) {
	ctx := context.Background()
	fv := field.ReflectValueOf(ctx, rv)
	if fv.IsValid() && fv.CanSet() {
		switch fv.Kind() {
		case reflect.Int, reflect.Int64:
			fv.SetInt(value)
		case reflect.Uint, reflect.Uint64:
			fv.SetUint(uint64(value))
		}
	}
}
