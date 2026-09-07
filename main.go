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


func middlewareLog(next http.Handler) http.Handler{

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		log.Printf("%v :: %v", r.Method, r.URL.Path)
		next.ServeHTTP(w,r)
	})
}

func middlewareLog2(next http.HandlerFunc) http.FuncHandler{
	return func(w http.ResponseWriter, r *http.Request){
		log.Printf(":v :: %v ", r.Method, r.URL.Path)
		next(w,r)
	}
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

	mux.Handle("/app/healthz", middlewareLog(http.HandlerFunc(readinessHandler)))
	// another option is creating the middleware using http.FuncHandler
	mux.Handle("/app/healthz2", middlewareLog2(readinessHandler))

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,  
	}
	err := server.ListenAndServe()
	if err != nil{
		log.Fatalf("THE ERROR: :%v", err)
	}
}
