package cli

import (
	"github.com/cmingxu/mpu/ai"
	"github.com/cmingxu/mpu/model"
	"github.com/cmingxu/mpu/server"

	cli2 "github.com/urfave/cli/v2"

	"github.com/rs/zerolog"
)

var commands = []*cli2.Command{
	{
		Name:  "server",
		Usage: "webserver for Money Printer Ultra",
		Flags: []cli2.Flag{
			&cli2.StringFlag{
				Name:  "listen-addr",
				Value: ":8080",
			},

			&cli2.StringFlag{
				Name:    "model",
				Value:   "gpt-3.5-turbo",
				EnvVars: []string{"MODEL"},
			},

			&cli2.StringFlag{
				Name:    "openai-key",
				Value:   "",
				EnvVars: []string{"OPENAI_KEY"},
			},

			&cli2.StringFlag{
				Name:    "openai-api",
				Value:   "",
				EnvVars: []string{"OPENAI_API"},
			},

			&cli2.StringFlag{
				Name:    "volengine-key",
				Value:   "",
				EnvVars: []string{"VOLENGINE_KEY"},
			},
		},

		Before: func(c *cli2.Context) error {
			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

			zeroLogLevel, err := zerolog.ParseLevel(c.String("log-level"))
			if err != nil {
				return err
			}

			zerolog.SetGlobalLevel(zeroLogLevel)

			return nil
		},

		Action: func(c *cli2.Context) error {
			if err := model.Init(c.String("work-dir") + "/mpu.db"); err != nil {
				return err
			}

			if err := model.InitDB(); err != nil {
				return err
			}

			ai.NewClient(c.String("model"),
				c.String("openai-key"),
				c.String("openai-api"))

			ai.NewTts(c.String("openai-key"))

			ai.NewTxt2Img(c.String("volengine-key"))

			s := server.New(c.String("listen-addr"), c.String("work-dir"))
			return s.Start()
		},
	},
}

var (
	logLeveLFlag = &cli2.StringFlag{
		Name:    "log-level",
		Value:   "info",
		EnvVars: []string{"LOG_LEVEL"},
	}

	workDirFlag = &cli2.StringFlag{
		Name:    "work-dir",
		Value:   "/data/mpu",
		EnvVars: []string{"WORK_DIR"},
	}

	dsnFlag = &cli2.StringFlag{
		Name: "dsn",
	}
)

func NewApp() *cli2.App {
	app := cli2.App{}
	app.Commands = commands

	app.Flags = []cli2.Flag{
		logLeveLFlag,
		workDirFlag,
	}

	app.Name = "Money Printer Ultra"
	app.Usage = "A CLI tool for printing money"

	return &app
}
