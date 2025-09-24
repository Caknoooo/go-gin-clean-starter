#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: ./create_module.sh <module_name>"
    exit 1
fi

MODULE_NAME=$1

PASCAL_MODULE_NAME=$(echo "$MODULE_NAME" | awk -F'_' '{for(i=1;i<=NF;i++){ $i=toupper(substr($i,1,1)) substr($i,2)} }1' OFS='')

CAMEL_MODULE_NAME="$(tr '[:upper:]' '[:lower:]' <<< ${PASCAL_MODULE_NAME:0:1})${PASCAL_MODULE_NAME:1}"

echo "🚀 Creating module: $MODULE_NAME -> $PASCAL_MODULE_NAME"
echo "📌 PascalCase: $PASCAL_MODULE_NAME | camelCase: $CAMEL_MODULE_NAME"

mkdir -p modules/$MODULE_NAME/controller
mkdir -p modules/$MODULE_NAME/service
mkdir -p modules/$MODULE_NAME/repository
mkdir -p modules/$MODULE_NAME/dto
mkdir -p modules/$MODULE_NAME/validation
mkdir -p modules/$MODULE_NAME/tests
mkdir -p modules/$MODULE_NAME/query

cat > modules/$MODULE_NAME/controller/${MODULE_NAME}_controller.go << EOF
package controller

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/$MODULE_NAME/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/$MODULE_NAME/validation"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type (
	${PASCAL_MODULE_NAME}Controller interface {
	}

	${CAMEL_MODULE_NAME}Controller struct {
		${CAMEL_MODULE_NAME}Service    service.${PASCAL_MODULE_NAME}Service
		${CAMEL_MODULE_NAME}Validation *validation.${PASCAL_MODULE_NAME}Validation
		db                             *gorm.DB
	}
)

func New${PASCAL_MODULE_NAME}Controller(injector *do.Injector, s service.${PASCAL_MODULE_NAME}Service) ${PASCAL_MODULE_NAME}Controller {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	${CAMEL_MODULE_NAME}Validation := validation.New${PASCAL_MODULE_NAME}Validation()
	return &${CAMEL_MODULE_NAME}Controller{
		${CAMEL_MODULE_NAME}Service:    s,
		${CAMEL_MODULE_NAME}Validation: ${CAMEL_MODULE_NAME}Validation,
		db:                             db,
	}
}
EOF

cat > modules/$MODULE_NAME/service/${MODULE_NAME}_service.go << EOF
package service

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/$MODULE_NAME/repository"
	"gorm.io/gorm"
)

type ${PASCAL_MODULE_NAME}Service interface {
}

type ${CAMEL_MODULE_NAME}Service struct {
	${CAMEL_MODULE_NAME}Repository repository.${PASCAL_MODULE_NAME}Repository
	db                            *gorm.DB
}

func New${PASCAL_MODULE_NAME}Service(
	${CAMEL_MODULE_NAME}Repo repository.${PASCAL_MODULE_NAME}Repository,
	db *gorm.DB,
) ${PASCAL_MODULE_NAME}Service {
	return &${CAMEL_MODULE_NAME}Service{
		${CAMEL_MODULE_NAME}Repository: ${CAMEL_MODULE_NAME}Repo,
		db:                            db,
	}
}
EOF

cat > modules/$MODULE_NAME/repository/${MODULE_NAME}_repository.go << EOF
package repository

import (
	"gorm.io/gorm"
)

type ${PASCAL_MODULE_NAME}Repository interface {
}

type ${CAMEL_MODULE_NAME}Repository struct {
	db *gorm.DB
}

func New${PASCAL_MODULE_NAME}Repository(db *gorm.DB) ${PASCAL_MODULE_NAME}Repository {
	return &${CAMEL_MODULE_NAME}Repository{
		db: db,
	}
}
EOF

cat > modules/$MODULE_NAME/dto/${MODULE_NAME}_dto.go << EOF
package dto

const (
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_SUCCESS_GET_DATA         = "success get data"
)

type (
	${PASCAL_MODULE_NAME}CreateRequest struct {
	}

	${PASCAL_MODULE_NAME}Response struct {
	}
)
EOF

cat > modules/$MODULE_NAME/validation/${MODULE_NAME}_validation.go << EOF
package validation

import (
	"github.com/go-playground/validator/v10"
)

type ${PASCAL_MODULE_NAME}Validation struct {
	validate *validator.Validate
}

func New${PASCAL_MODULE_NAME}Validation() *${PASCAL_MODULE_NAME}Validation {
	validate := validator.New()
	return &${PASCAL_MODULE_NAME}Validation{
		validate: validate,
	}
}
EOF

for file in controller service repository validation; do
cat > modules/$MODULE_NAME/tests/${MODULE_NAME}_${file}_test.go << EOF
package tests

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test${PASCAL_MODULE_NAME}$(echo $file | sed 's/.*/\u&/') (t *testing.T) {
	assert.True(t, true)
}
EOF
done

cat > modules/$MODULE_NAME/routes.go << EOF
package $MODULE_NAME

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/$MODULE_NAME/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	${CAMEL_MODULE_NAME}Controller := do.MustInvoke[controller.${PASCAL_MODULE_NAME}Controller](injector)

	${CAMEL_MODULE_NAME}Routes := server.Group("/api/$MODULE_NAME")
	{
		// TODO: add your endpoints here
	}
}
EOF

echo "✅ Module $PASCAL_MODULE_NAME created successfully!"
