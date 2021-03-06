package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/crypto/acme/autocert"

	"github.com/heartles/uttt/server/config"
	"github.com/heartles/uttt/server/socket"
	"github.com/heartles/uttt/server/store"
)

// ErrValidationFailed is returned if the path provided was not valid
var ErrValidationFailed = errors.New("path validation failed")

// ValidatePath checks that a user is allowed to access the given path.
// Note that, because of implementation details, it will deny any files
// which contain successive periods
func ValidatePath(path string) (string, error) {
	validated := filepath.Join("ui/", path)

	if strings.Contains(validated, "..") ||
		(strings.Index(validated, "ui/") != 0 && validated != "ui") {
		return "", ErrValidationFailed
	}

	return "./" + validated, nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config")
		panic(err)
	}

	server := buildServer(cfg)
	listenAddress := fmt.Sprintf(":%v", strconv.Itoa(cfg.Port))
	fmt.Printf("Listening on %v\n", listenAddress)
	if cfg.AcmeTLS {
		server.AutoTLSManager.HostPolicy = autocert.HostWhitelist("uttt.heartles.io")
		server.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")

		go server.Logger.Fatal(server.StartAutoTLS(listenAddress))

	} else {
		go server.Logger.Fatal(server.Start(listenAddress))
	}

	<-make(chan struct{})
}

// buildServer constructs an echo instance with the routes setup
// according to the configuration given
func buildServer(cfg *config.Config) *echo.Echo {
	server := echo.New()
	gameService, err := store.NewGameService(cfg.DBFilename)
	if err != nil {
		panic(err)
	}
	socketServer := socket.NewServer(cfg, gameService)

	server.Use(middleware.Recover())
	if cfg.RequestLogs {
		server.Use(middleware.Logger())
	}

	server.GET("/", func(e echo.Context) error {
		return e.Redirect(301, "/ui/index.html")
	})

	server.GET("/ui/*", func(e echo.Context) error {
		path := e.Request().URL.Path[len("/ui/"):]

		validatedPath, err := ValidatePath(path)
		if err != nil {
			return e.String(403, "Forbidden")
		}

		return e.File(validatedPath)
	})

	server.GET("/socket", func(e echo.Context) error {
		socketServer.Handle(e.Response(), e.Request())

		return nil
	})

	return server
}
