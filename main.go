package main

import (
	"fmt"
	"log"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r * http.Request){
	fmt.Fprintln(w, "Hello From server...")
}


func readinessHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("Server working..."))

}

func main(){
	fmt.Println("....")
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))
	image_fs := http.FileServer(http.Dir("./images"))	
	mux.HandleFunc("/app/bob", helloHandler)

	mux.Handle("/app/hello/", http.StripPrefix("/app/hello/", fs))
	mux.Handle("/app/public/", fs)
	
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", image_fs))

	mux.HandleFunc("/app/healthz", readinessHandler)

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,  
	}
	err := server.ListenAndServe()
	if err != nil{
		log.Fatalf("THE ERROR: :%v", err)
	}
}
