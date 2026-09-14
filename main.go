package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"


	"github.com/google/uuid"
	"github.com/itsllyaz/Chirpy/internal"
	"github.com/itsllyaz/Chirpy/internal/auth"
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
	DBURL string 
	JWTSECRETKEY []byte
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
	type NewUser struct{
		Password string `json:"password"`
		Email string `json:"email"`
	}
	type UserReponse struct { Email string `json:"email"` }
	var user NewUser 
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil{
		http.Error(w, "Invalid Request check your request please", http.StatusBadRequest)
		fmt.Println("The error is :: ", err)
		return 
	}
	if user.Email == "" || user.Password == ""{
		http.Error( w, "Can't Create User. Please Enter full information", http.StatusBadRequest)
		return 
	}
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil{
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return 
	}
	params := database.CreateUserParams{
		Password: hashedPassword,
		// Password: sql.NullString{String: hashedPassword, Valid: true},
		Email: user.Email,

		
	}
		
	_, err = cfg.dbQuries.CreateUser(r.Context(), params)
	if err != nil{
		http.Error(w, "Can't Create User", http.StatusInternalServerError)
		fmt.Println("CAN'T CREATE, ERROR:: ", err)
		return
	}
	user_reponse := UserReponse{Email:user.Email}
	json.NewEncoder(w).Encode(user_reponse)
}

func (cfg *apiConfig) createChirpy(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")

	var chirpy database.CreateChirpyParams 
	err := json.NewDecoder(r.Body).Decode(&chirpy)
	if err != nil{
		http.Error(w, "Invalid INfo...", http.StatusBadRequest)
		return
	}
	params := database.CreateChirpyParams{
    Body: chirpy.Body, 
		UserID: chirpy.UserID, 
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

func (cfg *apiConfig) userLogin(w http.ResponseWriter, r *http.Request){
	/*
		-login--- accept email and password  as string format
		-you accept password as string and hash it 
		-and compare and check it with the one in database 
		-if the password and email correct, you will return....  id - created_at - updated_at - email 
	*/
	type LoginRequest struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginRequest
	type User struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
		Token string `json:"token"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil{
		http.Error(w, "Invalid Format", http.StatusBadRequest)
		return
	}
	
	theUser, err := cfg.dbQuries.LoginUser(r.Context(), req.Email)
	if err != nil{
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		fmt.Println("THE ERROR: ", err)
		return
	}

	// jsonData, err := json.MarshalIndent(theUser, "", " ")
	// fmt.Println(string(jsonData))	
	//
	match, err := auth.CheckPasswordHash(req.Password, theUser.Password)
	if err != nil{
		http.Error(w, "Internal Problem", http.StatusInternalServerError)
		return
	}

	if !match{
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return 
	}
	
	tokenString, err := auth.MakeJWT(theUser.ID, theUser.Email, cfg.JWTSECRETKEY)
	if err != nil{
		http.Error(w, "Internal Error: ", http.StatusInternalServerError)
		fmt.Println("JWT INTERNAL ERROR: ", err)
		return 
	}
	user := User{
		ID: theUser.ID,
		CreatedAt: theUser.CreatedAt,
		UpdatedAt: theUser.UpdatedAt,
		Email: theUser.Email,
		Token: tokenString,
	}

	json.NewEncoder(w).Encode(user)
	fmt.Println("TOKEN: ", user.Token)

	// tokenString, err := auth.MakeJWT(user.ID, user.Email )
	// if err != nil{
	// 	fmt.Println("ERROR ERROR: ", err)
	//
	// }
	// incomingToken := fmt.Sprintf("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9%v",tokenString)
	// fmt.Println("THE TOKEN: ", tokenString)
	//
	// extractedData := &auth.CustomClaims{}
	// token, err := jwt.ParseWithClaims(incomingToken, extractedData, func(token *jwt.Token) (any, error){
	// 	return auth.JwtKey, nil
	// })
	//
	// if err != nil{
	// 	fmt.Println("error the token was fake or expired...")
	// }
	// if token.Valid{
	// 	fmt.Println("Success... server now who you are")
	// 	fmt.Printf("UserID: %v, and UserEmail: %v", extractedData.UserID, extractedData.Email)
	// }
	//
	// fmt.Println(".................")
	// fmt.Println(".................")
}

func (cfg *apiConfig) protectedHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "THIS IS PROTECTED PAGE. ONLY AUTHENTICATED USER CAN SEE IT!!")
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
	JWTSECRETKEY := []byte(os.Getenv("JWTSECRETKEY"))
	dbConn, err := sql.Open("postgres", dbURL)

	if err != nil{
		log.Printf("ERROR WHILE OPENING DB CONNECTION:: %v", err)
	}
	defer dbConn.Close()

	cfg := &apiConfig{
		dbQuries: database.New(dbConn),
		DBURL: dbURL,
		JWTSECRETKEY: JWTSECRETKEY,
	}
	
	mw := internal.New(cfg.JWTSECRETKEY)



	mux.Handle("/app/hello/", http.StripPrefix("/app/hello/", fs))
	mux.Handle("/app/public/", fs)
	
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", image_fs))

	// mux.Handle("/app/healthz", middlewareLog(http.HandlerFunc(readinessHandler)))
	mux.Handle("GET /api/healthz", middlewareLog(cfg.middlewareMetricsInc(http.HandlerFunc(readinessHandler))))
	// another option is creating the middleware using http.FuncHandler
	mux.Handle("/api/healthz2", middlewareLog2(readinessHandler))

	
	mux.Handle("GET /admin/metrics", http.HandlerFunc(cfg.writeAdminMetricsHandler)) 
	mux.Handle("POST /admin/reset", http.HandlerFunc(cfg.resetAdminMetricsHandler))

	mux.Handle("POST /api/users", http.HandlerFunc(cfg.createNewUser))

	mux.Handle("POST /admin/chirpys", http.HandlerFunc(cfg.createChirpy))
	mux.Handle("GET /api/chirpys", http.HandlerFunc(cfg.getChirpys))
	mux.Handle("GET /api/chirpys/{id}", http.HandlerFunc(cfg.getChirpyByID))

	mux.Handle("POST /api/login", http.HandlerFunc(cfg.userLogin))
	mux.Handle("GET /api/protected", mw.LoginAuthMiddleware(http.HandlerFunc(cfg.protectedHandler)))

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,  
	}
	err = server.ListenAndServe()
	if err != nil{
		log.Fatalf("THE ERROR: :%v", err)
	}


}
