package main

import (
	"fmt"
	"log"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r * http.Request){
	fmt.Fprintln(w, "Hello From server...")
}



func main(){
	fmt.Println("....")
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))
	mux.HandleFunc("/bob", helloHandler)
	mux.Handle("/hello", fs)
	server := http.Server{
		Addr: ":8080",
		Handler: mux,  
	}
	err := server.ListenAndServe()
	if err != nil{
		log.Fatalf("THE ERROR: :%v", err)
	}
}
