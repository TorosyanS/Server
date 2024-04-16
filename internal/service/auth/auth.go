package auth

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"test/internal/custom_errors"
	"test/internal/service/auth/entity"
	"time"
)

type Service struct {
	passwordSalt []byte
	tokenSalt    []byte

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	userStorage     map[string]string
	tokenStorage    map[string]string
}

func NewAuthService(passwordSalt, tokenSalt []byte) *Service {
	return &Service{
		passwordSalt:    passwordSalt,
		tokenSalt:       tokenSalt,
		accessTokenTTL:  time.Minute,
		refreshTokenTTL: 24 * time.Hour,
		userStorage:     make(map[string]string),
		tokenStorage:    make(map[string]string),
	}
}

func (a *Service) RegisterUser(login, password string) error {
	_, ok := a.userStorage[login]
	if ok {
		return custom_errors.ErrUserAlreadyExists
	}

	passwordHash := a.hashPassword(password)

	a.userStorage[login] = passwordHash

	fmt.Println(a.userStorage)

	return nil
}

func (a *Service) AuthUser(login, password string) (entity.Tokens, error) {
	passwordHash, ok := a.userStorage[login]
	if !ok {
		return entity.Tokens{}, custom_errors.ErrNotFound
	}

	isPasswordCorrect := a.doPasswordsMatch(passwordHash, password)
	if !isPasswordCorrect {
		return entity.Tokens{}, custom_errors.ErrIncorrectPassword
	}

	tokens, err := a.generateTokens(login)
	if err != nil {
		return tokens, err
	}

	return tokens, nil
}

func (a *Service) VerifyUser(token string) (string, error) {
	claims := &entity.AuthClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("incorrect method")
		}

		return a.tokenSalt, nil
	})
	if err != nil || !parsedToken.Valid {
		return "", fmt.Errorf("incorrect token: %v", err)
	}

	return claims.Login, nil
}

func (a *Service) RefreshToken(token string) (entity.Tokens, error) {
	claims := &entity.RefreshTokenClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("incorrect method")
		}

		return a.tokenSalt, nil
	})
	if err != nil || !parsedToken.Valid {
		return entity.Tokens{}, fmt.Errorf("incorrect refresh token: %v", err)
	}

	// поиск токена в хранилище
	login, ok := a.tokenStorage[claims.AccessTokenID]
	if !ok || login != claims.Login {
		return entity.Tokens{}, custom_errors.ErrNotFound
	}

	// валидация прошла успешно, можем генерить новую пару
	tokens, err := a.generateTokens(claims.Login)
	if err != nil {
		return tokens, err
	}

	// удаляем данные о старом токене, чтобы никто не мог дважды сгенерить новую пару
	delete(a.tokenStorage, claims.AccessTokenID)

	fmt.Println(a.tokenStorage)

	return tokens, nil
}

func (a *Service) generateTokens(login string) (entity.Tokens, error) {
	accessTokenID := uuid.NewString()
	accessToken, err := a.generateAccessToken(login)
	if err != nil {
		return entity.Tokens{}, err
	}
	// accessToken - для доступа
	refreshToken, err := a.generateRefreshToken(login, accessTokenID)
	if err != nil {
		return entity.Tokens{}, err
	}

	// добавляем ID нового токена в хранилище
	a.tokenStorage[accessTokenID] = login

	return entity.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *Service) generateAccessToken(login string) (string, error) {
	now := time.Now()
	claims := entity.AuthClaims{
		Login: login,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.accessTokenTTL)), // TTL - time to live
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(a.tokenSalt)
	if err != nil {
		return "", fmt.Errorf("token.SignedString: %w", err)
	}

	return signedToken, nil
}

func (a *Service) generateRefreshToken(login, accessTokenID string) (string, error) {
	now := time.Now()
	claims := entity.RefreshTokenClaims{
		Login:         login,
		AccessTokenID: accessTokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.refreshTokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(a.tokenSalt)
	if err != nil {
		return "", fmt.Errorf("token.SignedString: %w", err)
	}

	return signedToken, nil
}

func (a *Service) hashPassword(password string) string {
	var passwordBytes = []byte(password)
	var sha512Hasher = sha512.New()

	// 0123456789abcdef
	// 324324fab32473

	passwordBytes = append(passwordBytes, a.passwordSalt...)
	sha512Hasher.Write(passwordBytes)

	var hashedPasswordBytes = sha512Hasher.Sum(nil)
	//fmt.Printf("Хэш пароля: %s\n", string(hashedPasswordBytes))
	var hashedPasswordHex = hex.EncodeToString(hashedPasswordBytes)
	//fmt.Printf("Хэш пароля после энкода: %s\n", string(hashedPasswordHex))

	return hashedPasswordHex
}

func (a *Service) doPasswordsMatch(hashedPassword, currPassword string) bool {
	return hashedPassword == a.hashPassword(currPassword)
}
