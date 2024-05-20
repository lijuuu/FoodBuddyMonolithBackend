package controllers

import (
	"fmt"
	"foodbuddy/database"
	"foodbuddy/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUserList(c *gin.Context) {
	var user []model.User

	tx := database.DB.Select("*").Find(&user)
	if tx.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "failed to retrieve data from the database, or the data doesn't exists",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userlist": user,
	})
}

func GetBlockedUserList(c *gin.Context) {
	var blockedUsers []model.User

	tx := database.DB.Where("deleted_at IS NULL AND blocked =?", true).Find(&blockedUsers)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve data from the database, or the data doesn't exists", "ok": false,
		})
		return
	}

	if len(blockedUsers) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "no blocked users found", "ok": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"blockedUserList": blockedUsers,
		"ok":              true,
	})
}

func BlockUser(c *gin.Context) {
	var user model.User

	userIdStr := c.Param("userid")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
			"ok":    false,
		})
		return
	}

	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"ok":    false,
		})
		return
	}

	if user.Blocked {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"message": "user is already blocked",
			"user":    user,
			"ok":      true,
		})
		return
	}

	user.Blocked = true
	fmt.Println(user)
	tx := database.DB.Updates(&user)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to change the block status ",
			"ok":    false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "user is blocked", "user": user, "ok": true,
	})
}

func UnblockUser(c *gin.Context) {
	var user model.User

	userIdStr := c.Param("userid")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
			"ok":    false,
		})
		return
	}

	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
			"ok":    false,
		})
		return
	}

	if !user.Blocked {
		c.JSON(http.StatusAlreadyReported, gin.H{
			"message": "user is already unblocked",
			"user":    user,
			"ok":      true,
		})
		return
	}

	user.Blocked = false
	fmt.Println(user)
	tx := database.DB.Model(&user).UpdateColumn("blocked", false)
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to change the unblock status",
			"ok":    false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "user is unblocked", "user": user, "ok": true,
	})
}

func AdminLogin(c *gin.Context) {
	Email := c.PostForm("Email")
	Password := c.PostForm("Password")

	if Email == "" || Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid credentials, try again",
			"ok":    false,
		})
		return
	}

}
