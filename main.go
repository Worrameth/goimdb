package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/proullon/ramsql/driver"
)

type Movie struct {
	ID          int64   `json:"id"`
	ImdbID      string  `json:"imdbid"`
	Title       string  `json:"title"`
	Year        int     `json:"year"`
	Rating      float32 `json:"rating"`
	IsSuperHero bool    `json:"is_super_hero"`
}

func getAllMoviesHandler(c echo.Context) error {
	mvs := []Movie{}
	y := c.QueryParam("year")

	if y == "" {
		// Query from Database
		rows, err := db.Query(`SELECT * FROM goimdb`)
		if err != nil {
			log.Fatal("query error", err)
		}
		defer rows.Close()

		for rows.Next() {
			var m Movie
			if err := rows.Scan(&m.ID, &m.ImdbID, &m.Title, &m.Year, &m.Rating, &m.IsSuperHero); err != nil {
				return c.JSON(http.StatusInternalServerError, "scan:"+err.Error())
			}
			mvs = append(mvs, m)
		}
		if err := rows.Err(); err != nil {
			return c.JSON(http.StatusInternalServerError, "rows err"+err.Error())
		}
		return c.JSON(http.StatusOK, mvs)
	}

	year, err := strconv.Atoi(y)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	rows, err := db.Query(`
	SELECT * FROM goimdb WHERE year=?`, year)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.ImdbID, &m.Title, &m.Year, &m.Rating, &m.IsSuperHero); err != nil {
			return c.JSON(http.StatusInternalServerError, "scan:"+err.Error())
		}
		mvs = append(mvs, m)
	}
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mvs)
}

func getMoviebyIdHandler(c echo.Context) error {
	imdbID := c.Param("imdbID")

	row := db.QueryRow(`SELECT * FROM goimdb WHERE imdbID=?`, imdbID)
	var m Movie
	err := row.Scan(&m.ID, &m.ImdbID, &m.Title, &m.Year, &m.Rating, &m.IsSuperHero)
	switch err {
	case nil:
		return c.JSON(http.StatusOK, m)
	case sql.ErrNoRows:
		return c.JSON(http.StatusNotFound, map[string]string{"massage": "Not Found"})
	default:
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
}

func createMovieHandler(c echo.Context) error {
	mv := &Movie{}
	if err := c.Bind(mv); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	stmt, err := db.Prepare(`
	INSERT INTO goimdb (imdbID,title,year,rating,isSuperHero) VALUES (?,?,?,?,?)
	`)
	if err != nil {
		log.Fatal(http.StatusInternalServerError, err.Error())
	}
	defer stmt.Close()
	b := fmt.Sprintf("%v", mv.IsSuperHero)
	r, err := stmt.Exec(mv.ImdbID, mv.Title, mv.Year, mv.Rating, b)
	switch {
	case err == nil:
		id, _ := r.LastInsertId()
		mv.ID = id
		return c.JSON(http.StatusCreated, mv)
	case err.Error() == "UNIQUE constraint violation":
		return c.JSON(http.StatusConflict, "movie already exists")
	default:
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
}

var db *sql.DB

func conn() {
	var err error
	db, err = sql.Open("ramsql", "goimdb")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	conn()

	createTb := `
	CREATE TABLE IF NOT EXISTS goimdb (
	id INT AUTO_INCREMENT,
	imdbID TEXT NOT NULL UNIQUE,
	title TEXT NOT NULL,
	year INT NOT NULL,
	rating FLOAT NOT NULL,
	isSuperHero BOOLEAN NOT NULL,
	PRIMARY KEY (id)
	);
	`
	if _, err := db.Exec(createTb); err != nil {
		log.Fatal("create table error", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.GET("/movies", getAllMoviesHandler)
	e.GET("/movies/:imdbID", getMoviebyIdHandler)

	e.POST("/movies", createMovieHandler)

	port := "2565"
	log.Printf("starting... port:%s", port)

	log.Fatal(e.Start(":" + port))
}
