package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/itsllyaz/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

func middlewareLog2(next http.HandlerFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		log.Printf(":%v :: %v ", r.Method, r.URL.Path)
		next(w,r)
	}
}

type apiConfig struct{
	fileserverhits atomic.Int32
	dbQuries *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r * http.Request){
		cfg.fileserverhits.Add(1)
		
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) writeAdminMetricsHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","text/html")
	w.WriteHeader(http.StatusOK)
	const template = `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`
	fmt.Fprintf(w, template, &cfg.fileserverhits)
}

func (cfg *apiConfig) resetAdminMetricsHandler(w http.ResponseWriter, r *http.Request){
	cfg.fileserverhits.Store(0)	
	fmt.Fprintf(w, "RESTED TO 0, CURRENT VALUE: %v", &cfg.fileserverhits)
}

func (cfg *apiConfig) createNewUser(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	var user *database.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil{
		http.Error(w, "Invalid Request check your request please", http.StatusBadRequest)
		fmt.Println("The error is :: ", err)
		return 
	}
		
	createdUser, err := cfg.dbQuries.CreateUser(r.Context(), user.Email)
	if err != nil{
		http.Error(w, "Can't Create User", http.StatusInternalServerError)
		fmt.Println("CAN'T CREATE, ERROR:: ", err)
		return
	}
	json.NewEncoder(w).Encode(createdUser)
}

func main(){
	fmt.Println("....")
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))
	// cfg := &apiConfig{
	// 	dbQuries: database.New(dbConn),
	// }

		// database related....
	err := godotenv.Load()
	if err != nil{
		log.Fatal("Error loading .env file")
	}
	dbURL := os.Getenv("DB_URL")
	dbConn, err := sql.Open("postgres", dbURL)

	if err != nil{
		log.Printf("ERROR WHILE OPENING DB CONNECTION:: %v", err)
	}
	defer dbConn.Close()

	cfg := &apiConfig{
		dbQuries: database.New(dbConn),
	}



	image_fs := http.FileServer(http.Dir("./images"))	
	mux.HandleFunc("/app/bob", helloHandler)

	mux.Handle("/app/hello/", http.StripPrefix("/app/hello/", fs))
	mux.Handle("/app/public/", fs)
	
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", image_fs))

	// mux.Handle("/app/healthz", middlewareLog(http.HandlerFunc(readinessHandler)))
	mux.Handle("GET /api/healthz", middlewareLog(cfg.middlewareMetricsInc(http.HandlerFunc(readinessHandler))))
	// another option is creating the middleware using http.FuncHandler
	mux.Handle("/api/healthz2", middlewareLog2(readinessHandler))

	
	mux.Handle("GET /admin/metrics", http.HandlerFunc(cfg.writeAdminMetricsHandler)) 
	mux.Handle("POST /admin/reset", http.HandlerFunc(cfg.resetAdminMetricsHandler))

	mux.Handle("POST /admin/users", http.HandlerFunc(cfg.createNewUser))

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,  
	}
	err = server.ListenAndServe()
	if err != nil{
		log.Fatalf("THE ERROR: :%v", err)
	}


}
