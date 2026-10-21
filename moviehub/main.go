package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"
    "strings"

    "github.com/jackc/pgx/v5/pgxpool"
    _ "github.com/joho/godotenv/autoload"

    "moviehub/auth"
    "moviehub/handler"
    "moviehub/mw"
    "moviehub/store"
)

func main() {
    secret := os.Getenv("JWT_SECRET")
    if len(secret) < 32 {
        log.Fatal("JWT_SECRET ต้องยาวอย่างน้อย 32 ตัวอักษร")
    }
    pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()
    if err := pool.Ping(context.Background()); err != nil {
        log.Fatal("ต่อฐานข้อมูลไม่ติด ", err)
    }

    tickets := &auth.Issuer{Secret: []byte(secret), TTL: time.Hour}
    h := &handler.Handler{Store: &store.Store{Pool: pool}, Tickets: tickets}

    mux := http.NewServeMux()
    mux.HandleFunc("POST /api/auth/register", h.Register)
    mux.HandleFunc("POST /api/auth/login", h.Login)
    mux.HandleFunc("GET /api/me", login(h.Me))
    mux.HandleFunc("GET /api/genres", h.Genres)
    mux.HandleFunc("GET /api/movies", h.SearchMovies)
    mux.HandleFunc("GET /api/movies/{id}", h.GetMovie)
    mux.HandleFunc("GET /api/movies/{id}/reviews", h.ListReviews)

     port := os.Getenv("PORT") // Render ส่งมาให้ ในเครื่องไม่มีก็ใช้ 8080
    if port == "" {
        port = "8080"
    }
    origins := strings.Fields(os.Getenv("CORS_ORIGINS"))
   log.Println("MovieHub API ฟังอยู่ที่ :"+port, "ต้อนรับ", origins)
    log.Fatal(http.ListenAndServe(":"+port, mw.Logging(mw.CORS(origins, mux))))
}