package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	authHandler "test/internal/api/handler/auth"
	"test/internal/api/handler/get_value"
	"test/internal/api/handler/refresh_token"
	"test/internal/api/handler/register"
	"test/internal/api/handler/save_value"
	myMiddleware "test/internal/api/middleware"
	"test/internal/polymorphism/storage/map_storage"
	"test/internal/service/auth"
)

func main() {
	// .env должен быть в .gitignore в неучебных проектах
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	authService := auth.NewAuthService([]byte(os.Getenv("AUTH_SALT")), []byte(os.Getenv("TOKEN_SALT")))
	authMiddleware := myMiddleware.NewAuthMiddleware(authService)

	r := chi.NewRouter()

	//r.Use(myMiddleware.Timer)

	//r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(myMiddleware.JsonHeader)

	storage := map_storage.NewStorage()

	savePairHandler := save_value.NewHandler(storage)
	getValueHandler := get_value.NewHandler(storage)
	authorizationHandler := authHandler.NewHandler(authService)
	registerHandler := register.NewHandler(authService)
	refreshTokenHandler := refresh_token.NewHandler(authService)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.CheckToken)
		r.Method(http.MethodPost, "/save", savePairHandler)
		r.Method(http.MethodGet, "/find", getValueHandler)
	})

	r.Method(http.MethodPost, "/auth", authorizationHandler)
	r.Method(http.MethodPost, "/register", registerHandler)
	r.Method(http.MethodPost, "/refresh-token", refreshTokenHandler)

	err = http.ListenAndServe(":3000", r)
	if err != nil {
		panic("cannot create server")
	}
}

//func main() {
//	asymmetricEncryption()
//}

func asymmetricEncryption() {
	// user_id -> 1; pub_key -> 6237464abd; user_email -> example@bank.ru
	// генерируем ключи
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	//fmt.Printf("Приватный ключ: %s\n", hex.EncodeToString(privateKey))
	fmt.Printf("Публичный ключ: %s\n", hex.EncodeToString(publicKey))

	// сообщение которое мы договорились использовать для верификации
	message := []byte("example")

	// генерируем подпись
	// сторона фронта

	message2 := []byte("example")

	signature := ed25519.Sign(privateKey, message)
	fmt.Printf("Подпись: %s\n", hex.EncodeToString(signature))

	// публичный ключ + подпись + сообщение
	isValid := ed25519.Verify(publicKey, message2, signature)
	if !isValid {
		fmt.Printf("Подпись не верна\n")
		return
	}

	fmt.Printf("Подпись верна\n")
}
