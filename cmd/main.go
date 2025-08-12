package main

import (
	"context"
	"fmt"
	"github.com/alishashelby/marketplace/internal/application/middleware"
	"github.com/alishashelby/marketplace/internal/application/service"
	"github.com/alishashelby/marketplace/internal/application/validator"
	"github.com/alishashelby/marketplace/internal/infrastructure/repository/ad"
	"github.com/alishashelby/marketplace/internal/infrastructure/repository/user"
	"github.com/alishashelby/marketplace/internal/presenation/controller"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	postgresDB, err := connectPostgres()
	if err != nil {
		log.Print("PostgreSQL connection failed: ", err)
		return
	}
	defer postgresDB.Close() //nolint:errcheck

	log.Print("PostgreSQL connection established")

	client, err := connectMongo(ctx)
	if err != nil {
		log.Print("MongoDB connection failed: ", err)
		return
	}
	defer func() {
		if ok := client.Disconnect(ctx); ok != nil {
			log.Fatal(ok)
		}
	}()

	log.Println("MongoDB connection established")

	mongoDB := client.Database(os.Getenv("MONGO_DB"))

	handler, err := registerRoutes(postgresDB, mongoDB)
	if err != nil {
		log.Printf("Error initializing routes: %s", err)
		return
	}

	addr := ":" + os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("Listening on %s", addr)
	err = http.ListenAndServe(addr, handler)
	if err != nil {
		log.Print(err)
		return
	}
}

func getDotEnvVariable(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %s not set", key)
	}

	return value, nil
}

func connectPostgres() (*pgxpool.Pool, error) {
	userEnv, err := getDotEnvVariable("POSTGRES_USER")
	if err != nil {
		return nil, err
	}

	passwordEnv, err := getDotEnvVariable("POSTGRES_PASSWORD")
	if err != nil {
		return nil, err
	}

	dbNameEnv, err := getDotEnvVariable("POSTGRES_DB")
	if err != nil {
		return nil, err
	}

	db, err := pgxpool.New(context.Background(), fmt.Sprintf(
		"postgres://%s:%s@postgres:5432/%s",
		userEnv,
		passwordEnv,
		dbNameEnv,
	))
	if err != nil {
		return nil, err
	}

	if err = db.Ping(context.Background()); err != nil {
		db.Close()
		log.Println("Err:", err)
	}

	return db, nil
}

func connectMongo(ctx context.Context) (*mongo.Client, error) {
	userEnv, err := getDotEnvVariable("MONGO_USER")
	if err != nil {
		return nil, err
	}

	passwordEnv, err := getDotEnvVariable("MONGO_PASSWORD")
	if err != nil {
		return nil, err
	}

	dbNameEnv, err := getDotEnvVariable("MONGO_DB")
	if err != nil {
		return nil, err
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		fmt.Sprintf(
			"mongodb://%s:%s@marketplace_mongodb:27017/%s?authSource=admin",
			userEnv,
			passwordEnv,
			dbNameEnv,
		)))
	if err != nil {
		return nil, err
	}

	if err = client.Ping(context.Background(), nil); err != nil {
		return nil, err
	}

	return client, nil
}

func registerRoutes(postgresDB *pgxpool.Pool, mongoDB *mongo.Database) (http.Handler, error) {
	jwtService, err := service.NewJWTService()
	if err != nil {
		return nil, err
	}

	userRepo := user.NewUserRepoPostgres(postgresDB)
	userService := service.NewUserService(userRepo, jwtService)
	userValidator := validator.NewUserValidator()
	userController := controller.NewUserController(userService, userValidator)

	adRepo := ad.NewAdRepoMongoDB(mongoDB)
	adService := service.NewAdService(adRepo)
	adValidator := validator.NewAdValidator()
	adController := controller.NewAdController(adService, userService, adValidator)

	r := mux.NewRouter()

	public := r.NewRoute().Subrouter()

	authorized := r.NewRoute().Subrouter()
	authorized.Use(func(next http.Handler) http.Handler {
		return middleware.AuthMiddleware(jwtService, next)
	})

	public.HandleFunc("/api/register", userController.Register).Methods(http.MethodPost)
	public.HandleFunc("/api/login", userController.Login).Methods(http.MethodPost)
	public.HandleFunc("/api/ads", adController.GetAllAds).Methods(http.MethodGet)

	authorized.HandleFunc("/api/publish", adController.CreateAd).Methods(http.MethodPost)
	authorized.HandleFunc("/api/ads/", adController.GetAdsWithOwned).Methods(http.MethodGet)

	handler := middleware.LoggingMiddleware(r)
	handler = middleware.PanicMiddleware(handler)

	return handler, nil
}
