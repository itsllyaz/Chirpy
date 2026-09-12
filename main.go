package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	 "github.com/google/uuid"
	"github.com/itsllyaz/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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

func (cfg *apiConfig) createChirpy(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	type Chirpy struct{
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	var chirpy Chirpy 
	err := json.NewDecoder(r.Body).Decode(&chirpy)
	if err != nil{
		http.Error(w, "Invalid INfo...", http.StatusBadRequest)
		return
	}
	params := database.CreateChirpyParams{
    Body: sql.NullString{
        String: chirpy.Body,
        Valid:  true,
    },
    UserID: uuid.NullUUID{
        UUID: chirpy.UserID,
        Valid: true,
    },
}	
	createdChirpy, err := cfg.dbQuries.CreateChirpy(r.Context(), params)
	if err != nil{
		log.Fatal("Can't create chirpy")
		fmt.Println("DEBUG ACTION HERE:: ", err)
	}
	json.NewEncoder(w).Encode(createdChirpy)
}

func (cfg *apiConfig) getChirpys(w http.ResponseWriter, r *http.Request){
	// here baby...
	chirpys, err := cfg.dbQuries.GetChirpys(r.Context())
	if err != nil{
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		fmt.Println("THE ERROR: ", err)
		return 
	}
	jsonData, err := json.Marshal(chirpys)
	if err != nil{
		http.Error(w, "error marshaling json", http.StatusInternalServerError)
		return 
	}
	fmt.Fprintln(w, string(jsonData))
	
}

func (cfg *apiConfig) getChirpyByID(w http.ResponseWriter, r *http.Request){
	id :=  r.PathValue("id")
	parsedUUID, err := uuid.Parse(id)
	if err != nil{
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return 
	}
	chirpy, err := cfg.dbQuries.GetChirpyByID(r.Context(), parsedUUID)
	if err != nil{
		http.Error(w, "Can't Fetch data", http.StatusNotFound)
		fmt.Println("THE ERROR: ", err)
		return
	}
	jsonData, err := json.Marshal(chirpy)
	if err != nil{
		http.Error(w, "Internal problem", http.StatusInternalServerError)
		return 
	}
	fmt.Fprintln(w, string(jsonData))
	
}

func main(){
	fmt.Println("server running:....")
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("."))
	image_fs := http.FileServer(http.Dir("./images"))

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

	mux.Handle("POST /admin/chirpys", http.HandlerFunc(cfg.createChirpy))
	mux.Handle("GET /api/chirpys", http.HandlerFunc(cfg.getChirpys))
	mux.Handle("GET /api/chirpys/{id}", http.HandlerFunc(cfg.getChirpyByID))

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,  
	}
	err = server.ListenAndServe()
	if err != nil{
		log.Fatalf("THE ERROR: :%v", err)
	}


}
