package main

import (
	"context"
	"fmt"
	"github.com/educ-educ/gateway/internal/auth/auth_handlers"
	"github.com/educ-educ/gateway/internal/pkg/http_tools"
	"github.com/educ-educ/gateway/internal/pkg/server"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

// @title Gateway
// @version 1.0

const (
	Mb                   = 2 << 23
	RmqConnString string = "amqp://hljhaczs:J5W5ouprR8fKEkmGKvpmd0ijcs3BbX8J@toad.rmq.cloudamqp.com/hljhaczs"
)

func main() {
	err := godotenv.Load("deploy_gateway/.env")
	if err != nil {
		fmt.Print(err)
		return
	}

	// Zap logging
	logConfig := zap.NewDevelopmentConfig()
	logConfig.DisableStacktrace = true
	baseLogger, err := logConfig.Build()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer func() {
		if err = baseLogger.Sync(); err != nil {
			log.Fatalf("can't flush log entities: %v", err)
		}
	}()

	logger := baseLogger.Sugar()

	// Validation
	//validate := validator.New()

	// Routing
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(http_tools.ErrorsMiddleware(logger, 5*Mb))

	handlersRouter := router.Group("/auth_handlers")
	handlersRouter.Any("/*proxyPath", func(c *gin.Context) {
		remote, err := url.Parse("http://handlers_service:8000")
		if err != nil {
			panic(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.URL.Path = "/auth_handlers" + c.Param("proxyPath")
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		proxy.ServeHTTP(c.Writer, c.Request)
	})

	rmqContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	authRouter := router.Group("/auth")
	{
		conn, err := amqp.Dial(RmqConnString)
		defer func() {
			err = conn.Close()
			if err != nil {
				log.Print(err.Error())
			}
		}()

		roleRouter := authRouter.Group("/roles")
		{
			add := auth_handlers.NewAddRoleHandler(rmqContext, conn)
			roleRouter.POST("/add", add.Handle)

			getAll := auth_handlers.NewGetRolesHandler(rmqContext, conn)
			roleRouter.GET("/get-all", getAll.Handle)
		}

		userRouter := authRouter.Group("/users")
		{
			getAll := auth_handlers.NewGetUsersHandler(rmqContext, conn)
			userRouter.GET("/get-all", getAll.Handle)

			edit := auth_handlers.NewEditUserHandler(rmqContext, conn)
			userRouter.POST("/edit", edit.Handle)
		}

		register := auth_handlers.NewRegisterHandler(rmqContext, conn)
		authRouter.POST("/register", register.Handle)

		login := auth_handlers.NewLoginHandler(rmqContext, conn)
		authRouter.POST("/login", login.Handle)
	}

	addr := ":" + os.Getenv("SERVICE_PORT")
	serv := server.NewServer(logger, router, addr)
	err = serv.Start()
	if err != nil {
		logger.Fatal(err)
	}
}
