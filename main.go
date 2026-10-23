package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"os"
	
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Article struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Category    string    `json:"category"`
	Status      string    `json:"status"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
}

type ArticleInput struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

var db *sql.DB

// validate memeriksa aturan dari soal, mengembalikan pesan error (kosong kalau lolos)
func validate(in ArticleInput) string {
	if utf8.RuneCountInString(strings.TrimSpace(in.Title)) < 20 {
		return "title wajib diisi, minimal 20 karakter"
	}
	if utf8.RuneCountInString(strings.TrimSpace(in.Content)) < 200 {
		return "content wajib diisi, minimal 200 karakter"
	}
	if utf8.RuneCountInString(strings.TrimSpace(in.Category)) < 3 {
		return "category wajib diisi, minimal 3 karakter"
	}
	if in.Status != "publish" && in.Status != "draft" && in.Status != "thrash" {
		return "status harus publish, draft, atau thrash"
	}
	return ""
}

func main() {
	var err error
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// Pengaturan bawaan untuk Laragon lokal
		dsn = "root:@tcp(127.0.0.1:3306)/article?parseTime=true&multiStatements=true"
	}
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("gagal konek database: ", err)
	}

	// Jalankan migration otomatis
	driver, err := migratemysql.WithInstance(db, &migratemysql.Config{})
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "mysql", driver)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	r := gin.Default()
	r.Use(cors.Default()) // supaya frontend boleh memanggil API ini

	// Parameter pertama dinamai :id di dua route supaya Gin tidak bentrok.
	// Di route list, :id artinya limit.
	r.POST("/article/", createArticle)
	r.GET("/article/:id/:offset", listArticles)
	r.GET("/article/:id", getArticle)
	r.PUT("/article/:id", updateArticle)
	r.DELETE("/article/:id", deleteArticle)

	r.Run(":8080")
}

func createArticle(c *gin.Context) {
	var in ArticleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON tidak valid"})
		return
	}
	if msg := validate(in); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	_, err := db.Exec("INSERT INTO posts (title, content, category, status) VALUES (?, ?, ?, ?)",
		in.Title, in.Content, in.Category, in.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{})
}

func listArticles(c *gin.Context) {
	limit, err1 := strconv.Atoi(c.Param("id")) // :id di sini = limit
	offset, err2 := strconv.Atoi(c.Param("offset"))
	if err1 != nil || err2 != nil || limit < 0 || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit dan offset harus angka"})
		return
	}

	query := "SELECT id, title, content, category, status, created_date, updated_date FROM posts"
	args := []interface{}{}
	// Opsional: /article/10/0?status=publish (berguna untuk tab di frontend)
	if status := c.Query("status"); status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}
	query += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	articles := []Article{}
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Content, &a.Category, &a.Status, &a.CreatedDate, &a.UpdatedDate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		articles = append(articles, a)
	}
	c.JSON(http.StatusOK, articles)
}

func getArticle(c *gin.Context) {
	var a Article
	err := db.QueryRow("SELECT id, title, content, category, status, created_date, updated_date FROM posts WHERE id = ?",
		c.Param("id")).Scan(&a.ID, &a.Title, &a.Content, &a.Category, &a.Status, &a.CreatedDate, &a.UpdatedDate)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "article tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

func updateArticle(c *gin.Context) {
	var in ArticleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON tidak valid"})
		return
	}
	if msg := validate(in); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	var exists int
	if err := db.QueryRow("SELECT COUNT(*) FROM posts WHERE id = ?", c.Param("id")).Scan(&exists); err != nil || exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "article tidak ditemukan"})
		return
	}
	_, err := db.Exec("UPDATE posts SET title=?, content=?, category=?, status=? WHERE id=?",
		in.Title, in.Content, in.Category, in.Status, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func deleteArticle(c *gin.Context) {
	res, err := db.Exec("DELETE FROM posts WHERE id = ?", c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "article tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}