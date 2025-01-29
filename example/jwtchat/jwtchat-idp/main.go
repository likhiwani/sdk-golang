package main

import (
	"context"
	"log"
	"net/http"
	"ztna-core/sdk-golang/example/jwtchat/jwtchat-idp/exampleop"
	"ztna-core/sdk-golang/example/jwtchat/jwtchat-idp/storage"
)

func main() {
	ctx := context.Background()

	storage := storage.NewStorage(storage.NewUserStore())

	port := "9998"
	router := exampleop.SetupServer(ctx, "http://localhost:"+port, storage)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Printf("server listening on http://localhost:%s/", port)
	log.Println("press ctrl+c to stop")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
	<-ctx.Done()
}
