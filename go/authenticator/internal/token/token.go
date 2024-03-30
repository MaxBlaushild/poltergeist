package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"math/big"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	userIDKey = "userID"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrInvalidClaims = errors.New("invalid claims")
)

type client struct {
	privateKey *ecdsa.PrivateKey
}

type Client interface {
	New(userID uuid.UUID) (string, error)
	Verify(tokenString string) (*uuid.UUID, error)
}

func NewClient(hexPrivateKey string) (Client, error) {
	privateKey, err := hexToECDSAPrivateKey(hexPrivateKey)
	if err != nil {
		return nil, err
	}

	return &client{
		privateKey: privateKey,
	}, nil
}

func hexToECDSAPrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}

	privateKeyInt := new(big.Int).SetBytes(keyBytes)
	curve := elliptic.P256()
	x, y := curve.ScalarBaseMult(keyBytes)
	privateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: privateKeyInt,
	}

	return privateKey, nil
}

func (c *client) New(userID uuid.UUID) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodES256,
		jwt.MapClaims{
			userIDKey: userID.String(),
		})
	s, err := t.SignedString(c.privateKey)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (c *client) Verify(tokenString string) (*uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return &c.privateKey.PublicKey, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "jwt parse error")
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	maybeStringUseID, ok := claims[userIDKey]
	if !ok {
		return nil, ErrInvalidClaims
	}

	stringUserID, ok := maybeStringUseID.(string)
	if !ok {
		return nil, ErrInvalidClaims
	}

	userID, err := uuid.Parse(stringUserID)
	if err != nil {
		if !ok {
			return nil, ErrInvalidClaims
		}
	}

	return &userID, nil
}
