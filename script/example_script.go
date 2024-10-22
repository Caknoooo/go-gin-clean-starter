package script

import (
	"fmt"

	"gorm.io/gorm"
)

type (
	ExampleScript struct {
		db *gorm.DB
	}
)

func NewExampleScript(db *gorm.DB) *ExampleScript {
	return &ExampleScript{
		db: db,
	}
}

func (s *ExampleScript) Run() error {
	// your script here
	fmt.Println("example script running")
	return nil
}