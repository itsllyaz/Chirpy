package main

import (
	"fmt"
	"log"
	"net/http"
)

func main(){
	fmt.Println("....")

	server := http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
			fmt.Fprintln(w, "Hello from server...")
		}),
	}

	err := server.ListenAndServe()
	if err != nil{
		log.Fatalf("THE ERROR: :%v", err)
	}
}
