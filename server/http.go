package server

import (
	"net/http"

	"context"
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/trianglehasfoursides/bedroompop/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/tidwall/gjson"
	"github.com/trianglehasfoursides/bedroompop/consist"
	"github.com/trianglehasfoursides/bedroompop/database"
	"go.uber.org/zap"
)

func Auth(ctx *gin.Context) {
	username, password, ok := ctx.Request.BasicAuth()
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "can't authenticate",
		})
	}

	isValid := (username == config.Username) && (password == config.Password)
	if !isValid {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "not authorized",
		})
		return
	}

	ctx.Next()
}

func Start(ch chan os.Signal) {
	// router
	router := gin.Default()

	router.Use(Auth)

	// HTTP server
	server := &http.Server{
		Addr:    config.HTTPAddr,
		Handler: router,
	}

	go func() {
		<-ch
		server.Shutdown(context.TODO())
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zap.L().Sugar().Panic(err.Error())
	}
}

func create(ctx *gin.Context) {
	req := struct {
		Name      string `json:"name"`
		Migration string `json:"migration"`
	}{}

	if err := ctx.BindJSON(req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	if req.Name == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "name can't be empty",
		})
	}

	address := consist.Consist.LocateKey([]byte(req.Name)).String()
	if address == config.GRPCAddr {
		if err := database.Create(req.Name, req.Migration); err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
	}

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer conn.Close()
	client := NewPopServiceClient(conn)

	if _, err := client.Create(ctx, &RequestCreate{
		Name:      req.Name,
		Migration: req.Migration,
	}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
}

func get(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "name can't be empty",
		})
		return
	}

	address := consist.Consist.LocateKey([]byte(name)).String()
	if address == config.GRPCAddr {
		if err := database.Get(name); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer conn.Close()
	client := NewPopServiceClient(conn)

	if _, err := client.Get(ctx, &RequestGetDrop{
		Name: name,
	}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
}

func drop(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "name can't be empty",
		})
		return
	}

	address := consist.Consist.LocateKey([]byte(name)).String()
	if address == config.GRPCAddr {
		if err := database.Drop(name); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer conn.Close()
	client := NewPopServiceClient(conn)

	if _, err := client.Drop(ctx, &RequestGetDrop{
		Name: name,
	}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
}

func Query(ctx *gin.Context) {
	req := struct {
		Name  string `json:"name"`
		Query string `json:"query"`
	}{}

	if err := ctx.BindJSON(req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if req.Name == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "name can't be empty",
		})
		return
	}

	if req.Query == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "query can't be empty",
		})
		return
	}

	address := consist.Consist.LocateKey([]byte(req.Name)).String()
	if address == config.GRPCAddr {
		if _, err := database.Query(req.Name, req.Query); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer conn.Close()
	client := NewPopServiceClient(conn)

	if _, err := client.Query(ctx, &RequestQueryExec{
		Name:  req.Name,
		Query: req.Query,
	}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
}

func Exec(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	name, query, args := gjson.Get(string(body), "name").String(), gjson.Get(string(body), "query").String(), gjson.Get(string(body), "args").Array()
	if name == "" && query == "" {
		http.Error(w, "database name is required", 400)
		return
	}
	command(w, "exec", name, "", query, args...)
}
