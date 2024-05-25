package controllers

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"foodbuddy/database"
// 	model "foodbuddy/model"
// 	utils "foodbuddy/utils"
// 	"io"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"golang.org/x/oauth2"
// 	"golang.org/x/oauth2/google"
// 	"gorm.io/gorm"
// )

// var googleOauthConfig = &oauth2.Config{
// 	RedirectURL:  "http://localhost:8080/api/v1/googlecallback",
// 	ClientID:     utils.GetEnvVariables().ClientID,
// 	ClientSecret: utils.GetEnvVariables().ClientSecret,
// 	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
// 	Endpoint:     google.Endpoint,
// }

// func GoogleHandleLogin(c *gin.Context) {
// 	utils.NoCache(c)
// 	url := googleOauthConfig.AuthCodeURL("hjdfyuhadVFYU6781235")
// 	c.Redirect(http.StatusTemporaryRedirect, url)
// 	c.Next()
// }

// func GoogleHandleCallback(c *gin.Context) {
//     utils.NoCache(c)
//     fmt.Println("Starting to handle callback")
//     code := c.Query("code")
//     if code == "" {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "missing code parameter","ok": false,})
//         return
//     }

//     token, err := googleOauthConfig.Exchange(context.Background(), code)
//     if err!= nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token","ok": false,})
//         return
//     }

//     response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
//     if err!= nil {
//         fmt.Println("google signup done")
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info","ok": false,})
//         return
//     }
//     defer response.Body.Close()

//     content, err := io.ReadAll(response.Body)
//     if err!= nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read user info","ok": false,})
//         return
//     }

//     var User model.GoogleResponse
//     err = json.Unmarshal(content, &User)
//     if err!= nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info","ok": false,})
//         return
//     }

//     newUser := model.User{
//         Name:           User.Name,
//         Email:          User.Email,
//         LoginMethod:    model.GoogleSSOMethod,
//         Picture:        User.Picture,
//         VerificationStatus: model.VerificationStatusVerified,
//         Blocked: false,
//     }

//     if newUser.Name == ""{
//         newUser.Name = User.Email
//     }

//     // Check if the user already exists
//     var existingUser model.User
//     if err := database.DB.Where("email =? AND deleted_at IS NULL", newUser.Email).First(&existingUser).Error; err!= nil {
//         if err == gorm.ErrRecordNotFound {
//             // Create a new user
//             if err := database.DB.Create(&newUser).Error; err!= nil {
//                 c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create user using google signup method","ok": false})
//                 return
//             }
//         } else {
//             // Handle case where user already exists but not found due to other errors
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error while fetching user","ok": false})
//             return
//         }
//     }

//     // User already exists, check login method
//     if existingUser.LoginMethod == model.EmailLoginMethod {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists, please use email for login","ok": false})
//         return
//     }


//     //check is the user is blocked by the admin
//     if existingUser.Blocked{
//         c.JSON(http.StatusUnauthorized, gin.H{"error": "user is restricted from accessing, blocked by the administrator","ok": false})
//         return
//     }
    
//     // Generate JWT and set cookie within GenerateJWT
//     tokenstring := GenerateJWT(c, newUser.Email)
//     if tokenstring == ""{
//         c.JSON(http.StatusInternalServerError,gin.H{
//             "error":"jwt token is empty please try again",
//             "ok": false,
//         })
//         return
//     }

//     // Return success response
//     fmt.Println("google signup done")
//     c.JSON(http.StatusOK, gin.H{"message": "Logged in successfully", "user": existingUser,"jwttoken":tokenstring,"ok": true,})
    
// }