package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"foodbuddy/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetCategoryList(c *gin.Context) { //public
	var category []model.Category

	tx := database.DB.Select("*").Find(&category)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to retrieve data from the database, or the data doesn't exists",
			"error_code": http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{
		"status":  true,
		"message": "category list is fetched successfully",
		"data": gin.H{
			"categorylist": category,
		},
	})
}

func GetCategoryProductList(c *gin.Context) { //public
	var categories []model.Category
	if err := database.DB.Preload("Products").Find(&categories).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to retrieve data from the database, or the data doesn't exists",
			"error_code": http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "category list with products is fetched successfully",
		"data": gin.H{
			"categorylist": categories,
		},
	})
}

func AddCategory(c *gin.Context) { //admin

	var Request model.AddCategoryRequest
	var existingcategory model.Category

	//check admin api authentication
	_,role,err := utils.GetJWTClaim(c)
	if role != model.AdminRole || err != nil{
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "unauthorized request",
		})
		return
	}

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process incoming request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	//validate the struct body
	if err := validate(Request); err != nil {
		//add json response
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    err,
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	words := strings.Fields(Request.Description)
	wordCount := len(words)

	if wordCount < 10 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "description must be a minimum of 10 words",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	// Check if the category is already present
	if err := database.DB.Where("name =?", Request.Name).Find(&existingcategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to fetch information for possible category name match",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	if Request.Name == existingcategory.Name {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "category already exists",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if err := database.DB.Create(&Request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "unable to add new category, server error ",
			"error_code": http.StatusInternalServerError,
		})
		return

	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successully added a new category",
		"data": gin.H{
			"category": Request,
		},
	})
}
func EditCategory(c *gin.Context) { //admin

	var Request model.EditCategoryRequest
	var existingcategory model.Category

	//check admin api authentication
	_,role,err := utils.GetJWTClaim(c)
	if role != model.AdminRole || err != nil{
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "unauthorized request",
		})
		return
	}

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process incoming request",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if err := database.DB.First(&existingcategory, Request.ID).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to fetch category details from the database",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	if err := database.DB.Where("name =?", Request.Name).Find(&existingcategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to fetch category details via category name",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	Request.Name = existingcategory.Name

	if err := database.DB.Updates(&Request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to update category details",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully updated category",
	})
}

func DeleteCategory(c *gin.Context) { //admin

	var category model.Category

	catergoryIDStr := c.Param("categoryid")

	//check admin api authentication
	_,role,err := utils.GetJWTClaim(c)
	if role != model.AdminRole || err != nil{
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":     false,
			"message":    "unauthorized request",
		})
		return
	}

	categoryID, err := strconv.Atoi(catergoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "invalid category ID",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if err := database.DB.First(&category, categoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to fetch category from the database",
			"error_code": http.StatusNotFound,
		})
		return
	}

	if err := database.DB.Delete(&category, categoryID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to delete category from the database",
			"error_code": http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Successfully deleted category from the database",
	})
}
