package main

import (
	"context"
	"fmt"
	"log"
	"logger/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "2001"
	rpcPort  = "5001"
	mongoURL = "mongodb://admin:password@mongo:27017"
	grpcPort = "5001"
)

var client *mongo.Client

type Config struct {
	Models data.Model
}

func main() {
	mongoClient, err := mongoConnect()
	if err != nil {
		log.Printf("failed to create mongo client: %s", err.Error())
		return
	}
	log.Println("mongodb connected")
	client = mongoClient
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Printf("Failed to disconnect client: %s", err.Error())

		}
		log.Println("mongo client disconnected")
	}()

	log.Printf("Servie is coinfigured")
	app := &Config{
		Models: *data.New(client),
	}
	err = rpc.Register(new(RpcServer))
	if err != nil {
		log.Printf("Failed to register rpc server %s ", err.Error())
		return 
	}
	go app.startRpcServer()
	err = app.serve()
	if err != nil {
		log.Printf("Failed to start server: %s", err.Error())
		return
	}
}

func (app *Config) startRpcServer() {
	log.Println("RPC server is starting")
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		log.Printf("Failed to start rpc server")
		return
	}

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(rpcConn)
	}

}

func (app *Config) serve() error {
	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	log.Printf("logger-service starts: %s ", webPort)
	err := srv.ListenAndServe()
	if err != nil {
		log.Printf("Failed to start server")
		return err
	}

	return nil
}

func mongoConnect() (*mongo.Client, error) {
	ctx := context.TODO()
	client := options.Client().ApplyURI(mongoURL)
	conn, err := mongo.Connect(ctx, client)
	if err != nil {
		log.Printf("Failed to connect mongo Db : %s", err.Error())
		return nil, err
	}
	return conn, err
}
