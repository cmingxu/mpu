package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cmingxu/mpu/ai"
	"github.com/cmingxu/mpu/model"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Server struct {
	engine  *gin.Engine
	addr    string
	workdir string
}

func New(addr string, workdir string) *Server {
	s := &Server{
		addr:    addr,
		workdir: workdir,
		engine:  gin.Default(),
	}

	return s
}

func (s *Server) Start() error {
	log.Info().Msgf("Starting server on %s", s.addr)

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

	api.POST("/movies/:movie_id/generate_script", func(c *gin.Context) {
		moveieId := c.Param("movie_id")
		movieIdInt, err := strconv.Atoi(moveieId)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid movie ID"})
			return
		}

		movie, err := model.GetMovie(int64(movieIdInt))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		var binding struct {
			Idea string `json:"idea"`
		}

		if err := c.ShouldBindJSON(&binding); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		if movie.Idea.String != binding.Idea {
			movie.Idea = sql.NullString{String: binding.Idea, Valid: true}
			if err := movie.Update(); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})

				return
			}
		}

		scripts, err := ai.GetClient().GenerateScript(c, movie.Idea.String, time.Minute*3)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		canSerializeToScriptItems := false
		if err := json.Unmarshal([]byte(scripts), &model.MovieScript{}); err == nil {
			canSerializeToScriptItems = true
		}

		if !canSerializeToScriptItems {
			c.JSON(400, gin.H{"error": "Generated script is not in the correct format"})
		}

		movie.Script = sql.NullString{String: scripts, Valid: true}
		if err := movie.Update(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"data": movie})
	})

	api.POST("/movies/:movie_id/scripts/:scirpt_index/generate_voice", func(c *gin.Context) {
		moveieId := c.Param("movie_id")
		movieIdInt, err := strconv.Atoi(moveieId)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid movie ID"})
			return
		}

		movie, err := model.GetMovie(int64(movieIdInt))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		script, err := movie.GetScript()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		scriptIndex := c.Param("scirpt_index")
		scriptIndexInt, err := strconv.Atoi(scriptIndex)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid script index"})
			return
		}

		if scriptIndexInt < 0 || scriptIndexInt >= len(script.ScriptItems) {
			c.JSON(400, gin.H{"error": "Script index out of range"})
			return
		}

		item := script.ScriptItems[scriptIndexInt]
		content, err := ai.GetTTSInstance().GenerateAudio(c, item.ZhSubtitle)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		fp := fmt.Sprintf("movie/%d/audio/%d.mp3",
			movieIdInt, scriptIndexInt)
		if err := saveMp3File(content, s.workdir, fp); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		item.VoicePath = fp
		raw, _ := json.Marshal(script)
		movie.Script = sql.NullString{String: string(raw), Valid: true}
		if err := movie.Update(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"data": movie})
	})

	api.POST("/movies/:movie_id/generate_voice", func(c *gin.Context) {
		moveieId := c.Param("movie_id")
		movieIdInt, err := strconv.Atoi(moveieId)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid movie ID"})
			return
		}

		movie, err := model.GetMovie(int64(movieIdInt))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		script, err := movie.GetScript()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		for i, item := range script.ScriptItems {
			log.Info().Msgf("Generating voice for item %d: %s", i, item.ZhSubtitle)
			rawMp3, err := ai.GetTTSInstance().GenerateAudio(c, item.ZhSubtitle)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			fp := fmt.Sprintf("movie/%d/audio/%d.mp3",
				movieIdInt, i)

			if err := saveMp3File(rawMp3, s.workdir, fp); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			item.VoicePath = fp
		}

		raw, _ := json.Marshal(script)
		movie.Script = sql.NullString{String: string(raw), Valid: true}
		if err := movie.Update(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"data": movie})
	})

	api.POST("/movies/:movie_id/scripts/:scirpt_index/generate_image", func(c *gin.Context) {
		moveieId := c.Param("movie_id")
		movieIdInt, err := strconv.Atoi(moveieId)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid movie ID"})
			return
		}

		movie, err := model.GetMovie(int64(movieIdInt))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		script, err := movie.GetScript()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		scriptIndex := c.Param("scirpt_index")
		scriptIndexInt, err := strconv.Atoi(scriptIndex)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid script index"})
			return
		}

		if scriptIndexInt < 0 || scriptIndexInt >= len(script.ScriptItems) {
			c.JSON(400, gin.H{"error": "Script index out of range"})
			return
		}

		item := script.ScriptItems[scriptIndexInt]
		content, err := ai.GetTxt2Img().GenerateImage(c, item.ImagePrompt)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
		}

		fp := fmt.Sprintf("movie/%d/image/%d.png", movieIdInt, scriptIndexInt)
		if err := saveImageFile(content, s.workdir, fp); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		item.ImagePath = fp
		raw, _ := json.Marshal(script)
		movie.Script = sql.NullString{String: string(raw), Valid: true}
		if err := movie.Update(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"data": movie})
	})

	api.GET("voices_list", func(c *gin.Context) {
		c.JSON(200, gin.H{"data": ai.VoiceList})
	})

	api.POST("/movies/:movie_id/generate_image", func(c *gin.Context) {
		moveieId := c.Param("movie_id")
		movieIdInt, err := strconv.Atoi(moveieId)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid movie ID"})
			return
		}

		movie, err := model.GetMovie(int64(movieIdInt))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		script, err := movie.GetScript()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		for i, item := range script.ScriptItems {
			log.Info().Msgf("Generating voice for item %d: %s", i, item.ZhSubtitle)
			content, err := ai.GetTxt2Img().GenerateImage(c, item.ImagePrompt)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			fp := fmt.Sprintf("movie/%d/image/%d.png", movieIdInt, i)
			if err := saveImageFile(content, s.workdir, fp); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			item.ImagePath = fp
		}

		raw, _ := json.Marshal(script)
		movie.Script = sql.NullString{String: string(raw), Valid: true}
		if err := movie.Update(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"data": movie})
	})

	return s.engine.Run(s.addr)
}

func saveMp3File(content []byte, workdir, fp string) error {
	log.Info().Msgf("Saving mp3 file to %s", fp)

	abspath := filepath.Join(workdir, fp)

	if err := os.RemoveAll(abspath); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(abspath), 0777); err != nil {
		return err
	}

	if err := ioutil.WriteFile(abspath, content, 0777); err != nil {
		return err
	}

	return nil
}

func saveImageFile(content []byte, workdir, fp string) error {
	log.Info().Msgf("Saving image file to %s", fp)

	abspath := filepath.Join(workdir, fp)

	if err := os.RemoveAll(abspath); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(abspath), 0777); err != nil {
		return err
	}

	if err := ioutil.WriteFile(abspath, content, 0777); err != nil {
		return err
	}

	return nil
}
