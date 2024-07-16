package controllers

import (
	"foodbuddy/internal/database"
	"foodbuddy/internal/model"
	"foodbuddy/internal/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetCategoryList(c *gin.Context) { //public
	var categories []struct {
		ID              uint   `json:"id"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		ImageURL        string `json:"image_url"`
		OfferPercentage uint   `json:"offer_percentage"`
	}

	tx := database.DB.Model(&model.Category{}).Select("id, name, description, image_url, offer_percentage").Find(&categories)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to retrieve data from the database, or the data doesn't exist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "category list is fetched successfully",
		"data": gin.H{
			"categorylist": categories,
		},
	})
}

func GetCategoryProductList(c *gin.Context) { //public
	var categories []model.Category
	if err := database.DB.Preload("Products").Find(&categories).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "make sure products are added in their respective catgories inorder to be displayed here",
		})
		return
	}

	// Transform the categories slice to match the desired JSON structure
	transformedCategories := make([]map[string]interface{}, len(categories))
	for i, category := range categories {
		transformedCategory := map[string]interface{}{
			"id":               category.ID,
			"name":             category.Name,
			"description":      category.Description,
			"image_url":        category.ImageURL,
			"offer_percentage": category.OfferPercentage,
			"products":         []map[string]interface{}{},
		}

		for _, product := range category.Products {
			transformedProduct := map[string]interface{}{
				"id":               product.ID,
				"restaurant_id":    product.RestaurantID,
				"category_id":      product.CategoryID,
				"name":             product.Name,
				"description":      product.Description,
				"image_url":        product.ImageURL,
				"price":            product.Price,
				"preparation_time": product.PreparationTime,
				"max_stock":        product.MaxStock,
				"offer_amount":     product.OfferAmount,
				"stock_left":       product.StockLeft,
				"average_rating":   product.AverageRating,
				"veg":              product.Veg,
			}
			transformedCategory["products"] = append(transformedCategory["products"].([]map[string]interface{}), transformedProduct)
		}

		transformedCategories[i] = transformedCategory
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "category list with products is fetched successfully",
		"data": gin.H{
			"categorylist": transformedCategories,
		},
	})
}
func AddCategory(c *gin.Context) { //admin

	var Request model.AddCategoryRequest
	var existingcategory model.Category

	//check admin api authentication
	_, role, err := utils.GetJWTClaim(c)
	if role != model.AdminRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to process incoming request",
		})
		return
	}

	//validate the struct body
	if err := utils.Validate(Request); err != nil {
		//add json response
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": err,
		})
		return
	}

	words := strings.Fields(Request.Description)
	wordCount := len(words)

	if wordCount < 10 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "description must be a minimum of 10 words",
		})
		return
	}

	// Check if the category is already present
	if err := database.DB.Where("name = ?", Request.Name).First(&existingcategory).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "category already exists",
		})
		return
	}

	createCategory := model.Category{
		Name:        Request.Name,
		Description: Request.Description,
		ImageURL:    Request.ImageURL,
	}

	if err := database.DB.Save(&createCategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "unable to add new category, server error ",
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

func EditCategory(c *gin.Context) {
	var Request model.EditCategoryRequest
	var existingcategory model.Category

	// Check admin role
	_, role, err := utils.GetJWTClaim(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "error while validating admin role",
		})
		return
	}
	if role != model.AdminRole {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request: admin role required",
		})
		return
	}

	// Bind JSON request
	if err := c.BindJSON(&Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to process incoming request",
		})
		return
	}

	// Fetch existing category
	if err := database.DB.First(&existingcategory, Request.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "category not found",
		})
		return
	}

	// Update fields if changed
	if Request.Name != existingcategory.Name {
		existingcategory.Name = Request.Name
	}
	existingcategory.Description = Request.Description
	existingcategory.ImageURL = Request.ImageURL

	// Perform update
	if err := database.DB.Save(&existingcategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to update category details",
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully updated category",
	})
}

func DeleteCategory(c *gin.Context) { //admin

	var category model.Category

	catergoryIDStr := c.Query("categoryid")

	//check admin api authentication
	_, role, err := utils.GetJWTClaim(c)
	if role != model.AdminRole || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "unauthorized request",
		})
		return
	}

	categoryID, err := strconv.Atoi(catergoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "invalid category ID",
		})
		return
	}

	if err := database.DB.First(&category, categoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to fetch category from the database",
		})
		return
	}
	var productCount int64
	if result := database.DB.Model(&model.Product{}).Where("category_id = ?", categoryID).Count(&productCount); result.RowsAffected > 0 {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": false, "message": "category contains products, change the category of these products before using this endpoint"})
		return
	}

	if err := database.DB.Delete(&category, categoryID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to delete category from the database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Successfully deleted category from the database",
	})
}

func GetCategoryOfferfromProductID(ProductID uint) (bool, uint, uint) { //ok,id,offerpercentage
	var Product model.Product
	if err := database.DB.Where("id = ?", ProductID).First(&Product).Error; err != nil {
		return false, Product.CategoryID, 0
	}
	var Category model.Category
	if err := database.DB.Where("id = ?", Product.CategoryID).First(&Category).Error; err != nil {
		return false, Product.CategoryID, 0
	}

	return true, Product.CategoryID, Category.OfferPercentage
}

func GetOverallRestaurantOrderInfo(c *gin.Context) {

}
