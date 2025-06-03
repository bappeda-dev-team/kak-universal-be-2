package helper

import (
	"ekak_kabupaten_madiun/model/web"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
var jwtIssuer = os.Getenv("JWT_ISSUER")
var jwtExpiration = os.Getenv("JWT_EXPIRATION")

func CreateNewJWT(userId int, pegawaiId string, email string, nip string, kodeOpd string, namaPegawai string, roles []string) string {
	exp := 24 * time.Hour
	if jwtExpiration != "" {
		if duration, err := time.ParseDuration(jwtExpiration + "h"); err == nil {
			exp = duration
		}
	}

	claims := jwt.MapClaims{
		"iss":          jwtIssuer,
		"user_id":      userId,
		"pegawai_id":   pegawaiId,
		"email":        email,
		"nip":          nip,
		"kode_opd":     kodeOpd,
		"nama_pegawai": namaPegawai,
		"roles":        roles,
		"iat":          time.Now().Unix(),
		"exp":          time.Now().Add(exp).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecretKey)
	if err != nil {
		fmt.Println(err)
	}

	return signedToken
}

func ValidateJWT(tokenString string) web.JWTClaim {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		fmt.Println(err)
		return web.JWTClaim{}
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var roles []string
		if rolesInterface, exists := claims["roles"]; exists {
			if rolesArray, ok := rolesInterface.([]interface{}); ok {
				for _, role := range rolesArray {
					if roleStr, ok := role.(string); ok {
						roles = append(roles, roleStr)
					}
				}
			}
		}

		userId := 0
		if id, ok := claims["user_id"].(float64); ok {
			userId = int(id)
		}

		pegawaiId := ""
		if id, ok := claims["pegawai_id"].(string); ok {
			pegawaiId = id
		}

		email := ""
		if e, ok := claims["email"].(string); ok {
			email = e
		}

		nip := ""
		if n, ok := claims["nip"].(string); ok {
			nip = n
		}

		issuer := ""
		if iss, ok := claims["iss"].(string); ok {
			issuer = iss
		}

		iat := int64(0)
		if issuedAt, ok := claims["iat"].(float64); ok {
			iat = int64(issuedAt)
		}

		exp := int64(0)
		if expiry, ok := claims["exp"].(float64); ok {
			exp = int64(expiry)
		}

		kodeOpd := ""
		if opd, ok := claims["kode_opd"].(string); ok {
			kodeOpd = opd
		}

		return web.JWTClaim{
			Issuer:    issuer,
			UserId:    userId,
			PegawaiId: pegawaiId,
			KodeOpd:   kodeOpd,
			Email:     email,
			Nip:       nip,
			Roles:     roles,
			Iat:       iat,
			Exp:       exp,
		}
	}

	return web.JWTClaim{}
}
