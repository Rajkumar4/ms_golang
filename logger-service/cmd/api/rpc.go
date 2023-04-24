package main

import (
	"context"
	"fmt"
	"log"
	"logger/data"

	"github.com/google/uuid"
)

type RpcServer struct{}

type RpcPayload struct {
	Name string
	Data string
}

func (r *RpcServer) LogInfo(rpcpayload RpcPayload, resp *string) error {
	collection := client.Database("logs").Collection("logs")
	uid := uuid.New()
	log.Printf("check input payload %v",rpcpayload)
	_, err := collection.InsertOne(context.TODO(), data.Logger{
		ID:   uid.String(),
		Name: rpcpayload.Name,
		Data: rpcpayload.Data,
	})
	if err != nil {
		log.Printf("Failed to insert logs %s", err.Error())
		return err
	}
	temp := fmt.Sprintf("log is insert into database :%s", rpcpayload.Name)
	resp = &temp
	log.Printf("check the value %s",rpcpayload.Name)
	return nil
}
