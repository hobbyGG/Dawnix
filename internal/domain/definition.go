package domain

import "gorm.io/datatypes"

type ProcessDefinition struct {
	BaseModel

	Code      string
	Version   int
	Name      string
	Structure datatypes.JSON
	Config    datatypes.JSON
	IsActive  bool
}
