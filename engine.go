package ginger

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/METADIV-GO/ginger/pkg/logger"
	"github.com/METADIV-GO/gorm/conn"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	limiter "github.com/ulule/limiter/v3"
	_gin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"gorm.io/gorm"
)

/*
Engine is the framework's instance.
It is supposed to be one per application.
*/
var Engine = newEngine()

type engine struct {
	Gin             *gin.Engine  `json:"-"`
	DB              *gorm.DB     `json:"-"`
	MEM             *gorm.DB     `json:"-"`
	EnvironmentKeys []string     `json:"environment_keys"`
	Configs         engineConfig `json:"configs"`

	ApiHandlers  []ApiHandler        `json:"api_handlers"`
	WsHandlers   []WsHandler         `json:"ws_handlers"`
	CronHandlers []CornHandler       `json:"cron_handlers"`
	InitJobs     []InitJobHandler    `json:"init_jobs"`
	Middlewares  []MiddlewareHandler `json:"middlewares"`
}

type engineConfig struct {
	DBType  string
	MEMType string
}

func newEngine() *engine {
	return &engine{
		Gin:          gin.Default(),
		ApiHandlers:  make([]ApiHandler, 0),
		WsHandlers:   make([]WsHandler, 0),
		CronHandlers: make([]CornHandler, 0),
		InitJobs:     make([]InitJobHandler, 0),
		Middlewares:  make([]MiddlewareHandler, 0),
		EnvironmentKeys: []string{
			"GIN_MODE",
			"GIN_HOST",
			"GIN_PORT",
			"GORM_HOST",
			"GORM_PORT",
			"GORM_USERNAME",
			"GORM_PASSWORD",
			"GORM_DATABASE",
			"GORM_SILENT",
			"GORM_ENCRYPT_KEY",
		},
		Configs: engineConfig{
			DBType:  DB_TYPE_MYSQL,
			MEMType: DB_TYPE_MEM,
		},
	}
}

/*
LogErr logs the error message.
*/
func (e *engine) LogErr(msg ...any) {
	logger.ERROR(msg...)
}

/*
LogInfo logs the info message.
*/
func (e *engine) LogInfo(msg ...any) {
	logger.INFO(msg...)
}

/*
LogDebug logs the debug message.
*/
func (e *engine) LogDebug(msg ...any) {
	logger.DEBUG(msg...)
}

/*
SetDBType sets the database type.
*/
func (e *engine) SetDBType(dbType string) {
	e.Configs.DBType = dbType
}

/*
SetMEMType sets the memory type.
*/
func (e *engine) SetMEMType(memType string) {
	e.Configs.MEMType = memType
}

/*
GenerateTypescript generates typescript.
*/
func (e *engine) GenerateTypescript() {
	FileService.RemoveOldFiles()
	FileService.CreateFolder()
	ModelService.CreateModels()
	ApiService.CreateApis()
}

/*
Run starts the application.
*/
func (e *engine) Run() {
	e.setupDB()
	e.setupCors()
	e.executeBeforeJobs()
	e.registerApis()
	e.registerWs()
	e.registerCronJobs()
	e.executeAfterJobs()

	host := Env("GIN_HOST")
	port := Env("GIN_PORT")
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "5000"
	}
	e.Gin.Run(host + ":" + port)
}

func (e *engine) setupDB() {
	var silent bool
	if Env("GORM_SILENT") == "true" {
		silent = true
	} else {
		silent = Env("GIN_MODE") == "release"
	}

	var err error
	switch e.Configs.DBType {
	case DB_TYPE_MYSQL:
		e.DB, err = conn.QuickMySQL(silent)
		if err != nil {
			panic(err)
		}
	case DB_TYPE_PGSQL:
		e.DB, err = conn.QuickPostgreSQL(silent)
		if err != nil {
			panic(err)
		}
	case DB_TYPE_MEM:
		e.DB, err = conn.SqliteMem(silent)
		if err != nil {
			panic(err)
		}
	}
}

func (e *engine) setupCors() {
	allowOrigins := strings.Split(Env("CORS_ALLOW_ORIGINS"), ",")
	if len(allowOrigins) == 0 {
		allowOrigins = []string{"*"}
	}
	allowMethods := strings.Split(Env("CORS_ALLOW_METHODS"), ",")
	if len(allowMethods) == 0 {
		allowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	}
	allowHeaders := strings.Split(Env("CORS_ALLOW_HEADERS"), ",")
	if len(allowHeaders) == 0 {
		allowHeaders = []string{"Origin", "Authorization", "Content-Type", HEADER_X_LOCALE, HEADER_AUTHORIZATION}
	}

	e.Gin.Use(cors.New(cors.Config{
		AllowOrigins: allowOrigins,
		AllowMethods: allowMethods,
		AllowHeaders: allowHeaders,
	}))
}

func (e *engine) executeBeforeJobs() {
	for _, job := range e.InitJobs {
		if !job.After {
			job.Handler()
		}
	}
}

func (e *engine) executeAfterJobs() {
	for _, job := range e.InitJobs {
		if job.After {
			job.Handler()
		}
	}
}

func (e *engine) registerCronJobs() {
	cron := cron.New()
	for i := range e.CronHandlers {
		cron.AddFunc(e.CronHandlers[i].Pattern, e.CronHandlers[i].Handler)
		if e.CronHandlers[i].InitExec {
			e.CronHandlers[i].Handler()
		}
	}
	cron.Start()
}

func (e *engine) registerWs() {
	for _, ws := range e.WsHandlers {
		/*
			Request route
		*/
		route := strings.TrimRight(ws.Path, "/")
		handlers := make([]gin.HandlerFunc, 0)

		/*
			Middlewares
		*/
		for key, mid := range e.Middlewares {
			var skip = false
			for i := range mid.SkipPaths {
				if match, _ := regexp.Match(mid.SkipPaths[i], []byte(route)); match {
					skip = true
					break
				}
			}
			if skip {
				break
			}

			var match bool = false
			for i := range mid.MatchPaths {
				if match, _ = regexp.Match(mid.MatchPaths[i], []byte(route)); match {
					break
				}
			}
			if !match {
				continue
			}

			handlers = append([]gin.HandlerFunc{e.Middlewares[key].Handler}, handlers...)
		}

		e.Gin.GET(route, handlers...)
	}
}

func (e *engine) registerApis() {
	for _, api := range e.ApiHandlers {

		/*
			Request route
		*/
		route := strings.TrimRight(api.Path, "/")
		handlers := make([]gin.HandlerFunc, 0)

		/*
			Rate limit
		*/
		if api.Opts != nil && api.Opts.RateLimit != nil {
			handlers = append([]gin.HandlerFunc{
				_gin.NewMiddleware(limiter.New(memory.NewStore(), limiter.Rate{Period: api.Opts.RateLimit.Duration, Limit: api.Opts.RateLimit.Rate})),
			}, handlers...)
		}

		/*
			Middlewares
		*/
		for key, mid := range e.Middlewares {
			var skip = false
			for i := range mid.SkipPaths {
				if match, _ := regexp.Match(mid.SkipPaths[i], []byte(route)); match {
					skip = true
					break
				}
			}
			if skip {
				break
			}

			var match bool = false
			for i := range mid.MatchPaths {
				if match, _ = regexp.Match(mid.MatchPaths[i], []byte(route)); match {
					break
				}
			}
			if !match {
				continue
			}

			handlers = append([]gin.HandlerFunc{e.Middlewares[key].Handler}, handlers...)
		}

		/*
			Cache
		*/
		if api.Opts != nil && api.Opts.Cache != nil {
			handlers = append(handlers, cache.CachePage(persistence.NewInMemoryStore(time.Second), api.Opts.Cache.Duration, api.Handler))
		} else {
			handlers = append(handlers, api.Handler)
		}

		/*
			Methods
		*/
		switch api.Method {
		case http.MethodGet:
			e.Gin.GET(route, handlers...)
		case http.MethodPost:
			e.Gin.POST(route, handlers...)
		case http.MethodPut:
			e.Gin.PUT(route, handlers...)
		case http.MethodDelete:
			e.Gin.DELETE(route, handlers...)
		}
	}
}
