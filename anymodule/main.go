package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	moduleboilerplate "github.com/Twibbonize/go-module-boilerplate-mongodb"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	fiberadapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)


var mongoClient *mongo.Client
var redisClient redis.UniversalClient
var loggerMain         *slog.Logger

func initLogger() {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)

	loggerMain = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))
}

func ConstructResponse(
	c *fiber.Ctx,
	errorCode int,
	message string,
) error {

	var status bool
	if errorCode == fiber.StatusOK || errorCode == fiber.StatusCreated || errorCode == fiber.StatusAccepted || errorCode == fiber.StatusNoContent {
		status = true
	} else {
		status = false
	}

	return c.Status(errorCode).JSON(fiber.Map{
		"status": status,
		"message": message,
	})
}


func connectMongo() *mongo.Client {
	URI := os.Getenv("MONGODB_URI_SUBMISSION")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	client, errConnect := mongo.Connect(ctx, options.Client().ApplyURI(URI))

	if errConnect != nil {
		panic(errConnect)
	}

	if errPing := client.Ping(ctx, readpref.Primary()); errPing != nil {
		panic(errPing)
	}

	return client
}


func connectRedis() redis.UniversalClient {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		log.Fatal("REDIS_HOST environment variable not set")
	}

	if os.Getenv("APP_ENV") == "production" {
		clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    []string{redisHost},
			Password: os.Getenv("REDIS_PASS"),
		})

		_, err := clusterClient.Ping(context.Background()).Result()
		if err != nil {
			log.Fatal(err)
		}

		return clusterClient
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: os.Getenv("REDIS_PASS"),
		DB:       0, 
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}

	return client
}


var fiberLambda *fiberadapter.FiberLambda

// init the Fiber Server
func init() {
	initLogger()
	log.Printf("Fiber cold start")
	app := fiber.New()

	mongoClient = connectMongo()
	redisClient = connectRedis()

	anyCollection := mongoClient.Database("databaseName").Collection("moduleboilerplate")
	anyModuleSetter := moduleboilerplate.NewSetterLib(anyCollection, &redisClient)
	anyModuleGetter := moduleboilerplate.NewGetterLib(&redisClient)

	server := &server{
		anyModuleGetter: *anyModuleGetter,
		anyModuleSetter: *anyModuleSetter,
	}

	app.Get("/", func(c *fiber.Ctx) error {
		if os.Getenv("APP_ENV") == "development" {
			return c.SendString(fmt.Sprintf("/submission: %s", os.Getenv("APP_ENV")))
		}else{
			return c.SendString("Not Found")
		}
	})
	app.Get("seed-one-byrandid", func(c *fiber.Ctx) error {
		return server.SeedOneByRandId(c);
	})
	
	app.Get("seed-many", func(c *fiber.Ctx) error {
		return server.SeedMany(c);
	})
	
	app.Get("delete-many-byparticipant", func(c *fiber.Ctx) error {
		return server.DeleteManyByParticipant(c);
	})
	
	fiberLambda = fiberadapter.New(app)
}


// Handler will deal with Fiber working with Lambda
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req.Path = req.PathParameters["proxy"]
	return fiberLambda.ProxyWithContext(ctx, req)
}


func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(Handler)
}