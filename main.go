package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Movie struct {
	ImdbID      string  `json:"imdb_id"`
	Title       string  `json:"title"`
	Year        int     `json:"year"`
	Rating      float32 `json:"rating"`
	IsSuperHero bool    `json:"is_super_hero"`
}

var movies = []Movie{
	{
		ImdbID:      "tt4154796",
		Title:       "Avengers: Endgame",
		Year:        2019,
		Rating:      8.4,
		IsSuperHero: true,
	},
}

func getAllMoviesHandler(c echo.Context) error {
	y := c.QueryParam("year")

	if y == "" {
		return c.JSON(http.StatusOK, movies)
	}

	year, err := strconv.Atoi(y)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	ms := []Movie{}
	for _, m := range movies {
		if m.Year == year {
			ms = append(ms, m)
		}
	}

	return c.JSON(http.StatusOK, ms)
}

func getMoviebyIdHandler(c echo.Context) error {
	id := c.Param("id")

	for _, m := range movies {
		if m.ImdbID == id {
			return c.JSON(http.StatusOK, m)
		}
	}
	return c.JSON(http.StatusNotFound, map[string]string{"massage": "Not Found"})
}

func createMovieHandler(c echo.Context) error {
	mv := &Movie{}
	if err := c.Bind(mv); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	movies = append(movies, *mv)
	return c.JSON(http.StatusCreated, mv)
}

func main() {
	e := echo.New()

	e.GET("/movies", getAllMoviesHandler)
	e.GET("/movies/:id", getMoviebyIdHandler)

	e.POST("/movies", createMovieHandler)

	port := "2565"
	log.Printf("starting... port:%s", port)

	log.Fatal(e.Start(":" + port))
}
