package helper

import (
	"context"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/eyepatch5263/auth_jwt/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct{
	Email string
	First_name string
	Last_name string 
	Uid string
	User_type string
	jwt.StandardClaims
}

var userCollection *mongo.Collection=database.OpenCollection(database.Client,"user")

var SECRET_KEY string

func GenerateAllTokens(email string, firstName string, lastName string,userType string,uid string)(signedToken string, signedRefreshToken string,err error){
	claims:=&SignedDetails{
		Email: email,
		First_name: firstName,
		Last_name: lastName,
		Uid: uid,
		User_type: userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * 24 * 30).Unix(),
		},
	}
	refreshClaims:=&SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * 24 * 30*30).Unix(),
		},
	}

	token,err:=jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err!=nil{
		log.Panic(err)
		return
	}
	refreshToken,err:=jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err!=nil{
		log.Panic(err)
		return
	}
	return token,refreshToken,err
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string){
	ctx,cancel:=context.WithTimeout(context.Background(),100*time.Second)
	defer cancel()
	var updatedObj primitive.D
	updatedObj=append(updatedObj,bson.E{Key: "token", Value: signedToken})
	updatedObj=append(updatedObj,bson.E{Key:"refresh_token",Value:signedRefreshToken})

	updated_at,_:=time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
	updatedObj=append(updatedObj,bson.E{Key:"updated_at",Value:updated_at})
	upsert:=true
	filter:=bson.M{"user_id":userId}
	opt:=options.UpdateOptions{
		Upsert:&upsert,
	}
	_,err:=userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{Key:"$set",Value:updatedObj},
		},
		&opt,
	)

	if err!=nil{
		log.Panic(err)
		return
	}
}

func ValidateToken(signedToken string)(claims *SignedDetails,msg string){
	token,err:=jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token)(interface{},error){
			return []byte(SECRET_KEY),nil
		},
	)
	
	if err!=nil{
		msg=err.Error()
		return
	}
	claims,ok:=token.Claims.(*SignedDetails)
	if !ok{
		msg="Token is invalid"
		return
	}
	if claims.ExpiresAt<time.Now().Local().Unix(){
		msg="Token is expired"
		return
	}
	return claims,msg
}