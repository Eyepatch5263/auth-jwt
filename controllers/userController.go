package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/eyepatch5263/auth_jwt/database"
	helper "github.com/eyepatch5263/auth_jwt/helpers"
	"github.com/eyepatch5263/auth_jwt/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)


var userCollection *mongo.Collection=database.OpenCollection(database.Client,"user")

var validate=validator.New()
func HashPassword(password string) (string) {
	pass,err:=bcrypt.GenerateFromPassword([]byte(password),14)
	if err!=nil{
		log.Panic(err)
	}
	return string(pass)
}

func VerifyPassword(userPassword string, providedPassword string)(bool, string){
	err:=bcrypt.CompareHashAndPassword([]byte(providedPassword),[]byte(userPassword))
	check:=true
	msg:=""
	if err!=nil{
		msg = "email or password is incorrect"
		check=false
	}
	return check,msg

}
func Signup() gin.HandlerFunc{
	return func(c *gin.Context){
		ctx,cancel:=context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()
		var user models.User
		if err:=c.BindJSON(&user);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
			return
		}
		validationErr:=validate.Struct(user)
		if validationErr!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":validationErr.Error()})
			return
		}

		count,err:=userCollection.CountDocuments(ctx,bson.M{"email":user.Email})
		password:=HashPassword(*user.Password)
		user.Password=&password
		defer cancel()
		if err!=nil{
			log.Panic(err)
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occurred while checking for the email"})
		}
		if count>0{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Email or Phone already exists"})
			return
		}
		user.CreatedAt,_=time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		user.UpdatedAt,_=time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		user.ID=primitive.NewObjectID()
		user.User_id=user.ID.Hex()
		token,refreshToken,_:= helper.GenerateAllTokens(*user.Email,*user.First_name,*user.Last_name,*user.User_type,user.User_id)
		user.Token=&token
		user.Refresh_token=&refreshToken

		resultInsertionNumber,insertErr:=userCollection.InsertOne(ctx,user)
		if insertErr!=nil{
			msg:="User was not created"
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg})
			return
		}
		c.JSON(http.StatusOK,resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc{
	return func(c *gin.Context){
		ctx,cancel:=context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()
		var user models.User
		var foundUser models.User
		if err:=c.BindJSON(&user);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
			return
		}
		err:=userCollection.FindOne(ctx,bson.M{"email":user.Email}).Decode(&foundUser)
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"User not found"})
			return
		}
		passwordErr,_:=VerifyPassword(*user.Password,*foundUser.Password)
		if !passwordErr{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Incorrect email or password"})
			return
		}
		if foundUser.Email==nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"User not found"})
		}
		token,refreshToken,_:=helper.GenerateAllTokens(*foundUser.Email,*foundUser.First_name,*foundUser.Last_name,*foundUser.User_type,foundUser.User_id)
		helper.UpdateAllTokens(token,refreshToken,foundUser.User_id)

		if err:=userCollection.FindOne(ctx,bson.M{"user_id":foundUser.User_id}).Decode(&foundUser);err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
		}
		
		c.JSON(http.StatusOK,foundUser)
	}
}

func GetUsers()gin.HandlerFunc{
	return func(c *gin.Context) {
		if err:=helper.CheckUserType(c,"ADMIN");err!=nil{
			c.JSON(http.StatusUnauthorized,gin.H{"error":err.Error()})
			return
		}
		ctx,cancel:=context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()
		recordPerPage,err:= strconv.Atoi(c.Query("recordPerPage"))
		if err!=nil || recordPerPage<1{
			recordPerPage=10
		}
		page,err1:=strconv.Atoi(c.Query("page"))
		if err1!=nil || page<1{
			page=1
		}

		startIndex:=(page-1)*recordPerPage
		// startIndex,err=strconv.Atoi(c.Query("startIndex"))

		matchStage:=bson.D{{Key:"$match",Value:bson.D{{}}}}
		groupStage := bson.D{
			{Key:"$group", Value:bson.D{
				{Key:"_id", Value:nil},
				{Key:"total_count", Value:bson.M{"$sum": 1}},
				{Key:"data", Value:bson.M{"$push": "$$ROOT"}},
			}},
		}
		projectStage:=bson.D{
			{Key:"$project",Value:bson.D{
				{Key:"_id",Value:0},
				{Key:"total_count",Value:1},
				{Key:"user_items",Value:bson.D{{Key:"$slice",Value:[]interface{}{"$data",startIndex,recordPerPage}}}},
			}},
		}
		result,err:=userCollection.Aggregate(ctx,mongo.Pipeline{matchStage,groupStage,projectStage})
		if err!=nil{
			log.Panic(err)
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occurred while listing users"})
			return
		}
		var allUsers []bson.M
		if err=result.All(ctx,&allUsers);err!=nil{
			log.Fatal(err)
		}
		c.JSON(http.StatusOK,allUsers[0])
	}
}

func GetUser() gin.HandlerFunc{
	return func(c *gin.Context){
		userId:=c.Param("user_id")
		if err:=helper.MatchUserTypeToUid(c,userId);err!=nil{
			c.JSON(http.StatusUnauthorized,gin.H{"error":err.Error()})
			return
		}
		ctx,cancel:=context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()
		var user models.User
		err:=userCollection.FindOne(ctx,bson.M{"user_id":userId}).Decode(&user)
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"User not found"})
			return
		}
		c.JSON(http.StatusOK,user)
	}
}