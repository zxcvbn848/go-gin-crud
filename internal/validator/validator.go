package validator

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"
)

// HandleValidationError 處理驗證錯誤並返回友好的錯誤訊息
func HandleValidationError(c *gin.Context, err error) {
	var errors []string

	// 檢查是否為 validator.ValidationErrors 類型
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			errors = append(errors, getErrorMessage(fieldError))
		}
	} else {
		// 如果不是驗證錯誤，返回原始錯誤訊息
		errors = append(errors, err.Error())
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "驗證失敗",
		"details": errors,
	})
}

// getErrorMessage 根據驗證標籤返回友好的錯誤訊息
func getErrorMessage(fieldError validator.FieldError) string {
	fieldName := getFieldName(fieldError.Field())

	switch fieldError.Tag() {
	case "required":
		return fieldName + " 為必填欄位"
	case "email":
		return fieldName + " 格式不正確"
	case "min":
		return fieldName + " 長度或數值不能小於 " + fieldError.Param()
	case "max":
		return fieldName + " 長度或數值不能大於 " + fieldError.Param()
	case "oneof":
		return fieldName + " 必須是以下值之一: " + strings.ReplaceAll(fieldError.Param(), " ", ", ")
	case "omitempty":
		return fieldName + " 驗證失敗"
	case "gte":
		return fieldName + " 必須大於或等於 " + fieldError.Param()
	case "lte":
		return fieldName + " 必須小於或等於 " + fieldError.Param()
	case "gt":
		return fieldName + " 必須大於 " + fieldError.Param()
	case "lt":
		return fieldName + " 必須小於 " + fieldError.Param()
	default:
		return fieldName + " 驗證失敗: " + fieldError.Tag()
	}
}

// getFieldName 將欄位名稱轉換為友好的中文名稱
func getFieldName(field string) string {
	fieldMap := map[string]string{
		"Email":        "電子郵件",
		"Password":     "密碼",
		"Role":         "角色",
		"Title":        "標題",
		"Author":       "作者",
		"Name":         "名稱",
		"Description":  "描述",
		"Price":        "價格",
		"Stock":        "庫存",
		"Content":      "內容",
		"Page":         "頁碼",
		"PageSize":     "每頁筆數",
		"Search":       "搜尋關鍵字",
		"RefreshToken": "刷新令牌",
	}

	if name, ok := fieldMap[field]; ok {
		return name
	}
	return field
}
