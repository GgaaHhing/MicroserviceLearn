package jwt_op

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"microserviceLearn/microservice/log"
	"microserviceLearn/microservice_part3/internal"
	"time"
)

const (
	TokenExpired     = "Token已过期"
	TokenNotValueYet = "Token不再有效"
	TokenMalformed   = "Token非法"
	TokenInvalid     = "Token无效"
)

type CustomClaims struct {
	ID int32 //用户ID
	//添加JWt的标准结构，用以使用jwt
	jwt.StandardClaims
	//添加自己需要组合的字段
	NickName    string
	AuthorityId int32 //权限
}

type JWT struct {
	SigningKey []byte
}

func NewJwt() *JWT {
	return &JWT{SigningKey: []byte(internal.AppConf.JWTConfig.SigningKey)}
}

// GenerateToken 生成JWTToken
func (j *JWT) GenerateToken(claims CustomClaims) (string, error) {
	//第一个参数：jwt加密方式，第二个参数，一个带有jwt标准结构体的自定义结构体
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//传入一个专属的密钥
	tokenStr, err := token.SignedString(j.SigningKey)
	if err != nil {
		log.Logger.Error("生成JWTToken错误：" + err.Error())
		return "", err
	}
	return tokenStr, nil
}

// ParseToken 解析Token
func (j *JWT) ParseToken(tokenStr string) (*CustomClaims, error) {
	// keyFUnc的目的是验证密钥，你可以在这里进行正确的加密或解密处理，然后你最终return的必须是正确的
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		var result jwt.ValidationError
		if errors.As(err, &result) {
			switch result.Errors {
			case jwt.ValidationErrorExpired:
				return nil, errors.New(TokenExpired)
			case jwt.ValidationErrorNotValidYet:
				return nil, errors.New(TokenNotValueYet)
			case jwt.ValidationErrorMalformed:
				return nil, errors.New(TokenMalformed)
			default:
				return nil, errors.New(TokenInvalid)
			}
		}
	}

	if token == nil || !token.Valid {
		return nil, errors.New(TokenInvalid)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New(TokenInvalid)
	}

	return claims, nil
}

// RefreshToken 刷新Token
func (j *JWT) RefreshToken(tokenStr string) (string, error) {
	// TimeFunc 在解析令牌以验证 “exp ”声明（过期时间）时提供当前时间。
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}
	//token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
	//	return j.SigningKey, nil
	//})
	//if err != nil {
	//	return "", err
	//}
	//
	//if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
	//	jwt.TimeFunc = time.Now
	//	//增添时间
	//	claims.StandardClaims.ExpiresAt = time.Now().Add(7 * 24 * time.Hour).Unix()
	//	return j.GenerateToken(*claims)
	//}
	//return "", TokenInvalid
	claims, err := j.ParseToken(tokenStr)
	if err != nil {
		return "", errors.New(TokenInvalid)
	}
	claims.StandardClaims.ExpiresAt = time.Now().Add(time.Hour * 24).Unix()
	return j.GenerateToken(*claims)
}
