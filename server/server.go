package server

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/cmingxu/mpu/model"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	engine *gin.Engine
	addr   string
}

func New(addr string) *Server {
	s := &Server{
		addr:   addr,
		engine: gin.Default(),
	}

	return s
}

func (s *Server) Start() error {
	log.Infof("Starting server on %s", s.addr)

	s.engine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})

	api := s.engine.Group("/api")
	api.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": "1.0.0"})
	})

	api.GET("/templates", func(c *gin.Context) {
		list, err := model.ListTemplates()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"data": list})
	})

	api.GET("/templates/:id", func(c *gin.Context) {
		id := c.Param("id")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		}

		if tpl, err := model.GetTemplate(int64(idInt)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"data": tpl})
		}
	})

	api.GET("/templates/default", func(c *gin.Context) {
		d, err := model.DefaultTemplates()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"data": d})
	})

	api.POST("/movies", func(c *gin.Context) {
		var binding struct {
			Idea    string `json:"idea"`
			TplName string `json:"tpl_name"`
		}

		if err := c.ShouldBindJSON(&binding); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		movie := model.NewMovie()
		movie.Idea = sql.NullString{String: binding.Idea, Valid: true}

		_, err := model.GetTemplateByName(binding.TplName)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid template name"})
			return
		}

		if err := movie.Create(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(201, gin.H{"message": "Movie created successfully", "data": movie})
	})

	api.GET("/movies", func(c *gin.Context) {
		if movies, err := model.ListMovies(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
		} else {
			count, err := model.MoviesCount()
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"data": movies, "total": count})
		}
	})

	api.GET("/movies/:movie_id", func(c *gin.Context) {
		movieId := c.Param("movie_id")
		movieIdInt, err := strconv.Atoi(movieId)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid movie ID"})
			return
		}

		if movie, err := model.GetMovie(int64(movieIdInt)); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
		} else {
			c.JSON(200, gin.H{"data": movie})
		}
	})

	return s.engine.Run(s.addr)
}
