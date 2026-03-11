package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	UID      string `json:"uid"`
	Username string `json:"username"`
	IssuedAt int64  `json:"iat"`
	ExpireAt int64  `json:"exp"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

var jwtEncoding = base64.RawURLEncoding

func GenerateJWT(signingKey string, claims JWTClaims) (string, error) {
	if signingKey == "" {
		return "", errors.New("jwt signing key is empty")
	}
	if claims.UID == "" || claims.Username == "" {
		return "", errors.New("jwt claims uid/username is required")
	}
	if claims.IssuedAt == 0 {
		claims.IssuedAt = time.Now().Unix()
	}
	if claims.ExpireAt == 0 {
		claims.ExpireAt = time.Now().Add(7 * 24 * time.Hour).Unix()
	}

	headerJSON, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerPart := jwtEncoding.EncodeToString(headerJSON)
	payloadPart := jwtEncoding.EncodeToString(payloadJSON)
	signingInput := headerPart + "." + payloadPart

	sig := signHS256(signingKey, signingInput)
	return signingInput + "." + jwtEncoding.EncodeToString(sig), nil
}

func ParseAndValidateJWT(signingKey string, token string) (*JWTClaims, error) {
	if signingKey == "" {
		return nil, errors.New("jwt signing key is empty")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid jwt format")
	}

	signingInput := parts[0] + "." + parts[1]
	wantSig := signHS256(signingKey, signingInput)
	gotSig, err := jwtEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid jwt signature encoding")
	}
	if !hmac.Equal(wantSig, gotSig) {
		return nil, errors.New("invalid jwt signature")
	}

	payloadBytes, err := jwtEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid jwt payload encoding")
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errors.New("invalid jwt payload json")
	}
	now := time.Now().Unix()
	if claims.ExpireAt != 0 && now > claims.ExpireAt {
		return nil, errors.New("jwt expired")
	}
	if claims.UID == "" || claims.Username == "" {
		return nil, errors.New("jwt claims missing uid/username")
	}
	return &claims, nil
}

func signHS256(signingKey string, signingInput string) []byte {
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(signingInput))
	return mac.Sum(nil)
}
