package http

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
	"xy_im/internal/logic/conf"
)

var Srv http.Server

func New(c *conf.Config) {
	gin.SetMode(c.Env.DeployEnv)
	Srv = http.Server{
		Addr:         c.HTTPServer.Addr,
		Handler:      initRouter(),
		ReadTimeout:  time.Duration(c.HTTPServer.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(c.HTTPServer.WriteTimeout) * time.Second,
	}
	go func() {
		if err := Srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

}

func Close() {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := Srv.Shutdown(c); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
