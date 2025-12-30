package repository

import (
	"errors"
	"fmt"
	"runtime/debug"

	"case-trading/app/helper/database"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func GetService() *Service {
	return &Service{
		DB: database.GetDB(),
	}
}
func GetTransaction() *Service {
	tx := database.GetDB().Begin()
	fmt.Println("begin...")
	return &Service{
		DB: tx,
	}
}

func (s *Service) Commit() error {
	fmt.Println("commit...")
	return s.DB.Commit().Error
}

func (s *Service) Rollback(err ...interface{}) error {
	s.DB.Rollback()
	fmt.Println("rollback...")
	fmt.Println("panic error:", err)
	debug.PrintStack()

	if len(err) > 0 {
		if e, ok := err[0].(error); ok {
			return e
		}
	}

	return errors.New("transaction rollback")
}
