package controllers

import (
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetCategoryList(c *gin.Context) {
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

func GetCategoryProductList(c *gin.Context) {
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

func AddCategory(c *gin.Context) {


	var category model.Category
	var existingcategory model.Category

	if err := c.BindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process incoming request",
			"error_code": http.StatusBadRequest,
			
		})
		return
	}

	//validate the struct body
	if err := validate(category); err != nil {
		//add json response
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to validate category information",
			"error_code": http.StatusInternalServerError,
			
		})
		return
	}

	// Check if the category is already present
	if err := database.DB.Where("name =?", category.Name).Find(&existingcategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to fetch information for possible category name match",
			"error_code": http.StatusInternalServerError,
			
		})
		return
	}

	if category.Name == existingcategory.Name {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "category already exists",
			"error_code": http.StatusBadRequest,
			
		})
		return
	}

	category.ID = 0
	if err := database.DB.Create(&category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "unable to add new category, server error ",
			"error_code": http.StatusInternalServerError,
			
		})
		return
		
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"message":    "successully added a new category",
		"data":       gin.H{
			"category":category,
		},
	})
}
func EditCategory(c *gin.Context) {


	var category model.Category
	var existingcategory model.Category

	if err := c.BindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "failed to process incoming request",
			"error_code": http.StatusBadRequest,
			
		})
		return
	}

	if err := database.DB.First(&existingcategory, category.ID).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to fetch category details from the database",
			"error_code": http.StatusInternalServerError,
			
		})
		return
	}

	if err := database.DB.Where("name =?", category.Name).Find(&existingcategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to fetch category details via category name",
			"error_code": http.StatusInternalServerError,
			
		})
		return
	}

	category.Name = existingcategory.Name

	if err := database.DB.Updates(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     false,
			"message":    "failed to update category details",
			"error_code": http.StatusInternalServerError,
			
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"message":    "successfully updated category",
		
	})
}

func DeleteCategory(c *gin.Context) {

	var category model.Category

	catergoryIDStr := c.Param("categoryid")

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
		"status":     true,
		"message":    "Successfully deleted category from the database",
	})
}
