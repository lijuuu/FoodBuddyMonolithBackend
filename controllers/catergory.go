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
			"error": "failed to retrieve data from the database, or the data doesn't exists",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categorylist": category,
		"ok":           true,
	})
}

func AddCategory(c *gin.Context) {

	var category model.Category
	var existingcategory model.Category

	if err := c.BindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//validate the struct body
	// validate := validator.New()
	// err := validate.Struct(category)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"error": "failed to validate the struct body",
	// 		"ok":    false,
	// 	})
	// 	return
	// }

	// Check if the category is already present
	if err := database.DB.Where("name =?", category.Name).Find(&existingcategory).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error while checking category name",
			"ok":    false,
		})
		return
	}

	if category.Name == existingcategory.Name {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "category already exists",
			"ok":    false,
		})
		return
	}

	category.ID = 0
	if err := database.DB.Create(&category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":        "unable to add new category, possibly server error ",
			"errordetails": err,
			"ok":           false,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully added a new catergory",
		"ok":      true,
	})
}
func EditCategory(c *gin.Context) {

	var category model.Category
	var existingcategory model.Category

	if err := c.BindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error while binding json",
			"ok":    true,
		})
		return
	}

	if err := database.DB.First(&existingcategory, category.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch category details from the database",
			"ok":    true,
		})
		return
	}

	if err := database.DB.Where("name =?", category.Name).Find(&existingcategory).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error while checking category name",
			"ok":    false,
		})
		return
	}

	if category.Name == existingcategory.Name {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "category already exists",
			"ok":    false,
		})
		return
	}


	if err := database.DB.Updates(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update category",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "category updated successfully",
		"ok":      true,
	})
}

func DeleteCategory(c *gin.Context) {

	var category model.Category

	catergoryIDStr := c.Param("categoryid")

	categoryID, err := strconv.Atoi(catergoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid category ID",
			"ok":    false,
		})
		return
	}

	if err := database.DB.First(&category, categoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "category is not present on the database",
			"ok":    false,
		})
		return
	}

	if err := database.DB.Delete(&category, categoryID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to delete the category from the database",
			"ok":    false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully  deleted the category",
		"ok":      true,
	})

}
