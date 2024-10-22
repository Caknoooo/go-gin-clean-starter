package script

import (
	"errors"

	"gorm.io/gorm"
)

func Script(scriptName string, db *gorm.DB) error {
	switch scriptName {
	case "example_script":
		exampleScript := NewExampleScript(db)
		return exampleScript.Run()
	default:
		return errors.New("script not found")
	}
}