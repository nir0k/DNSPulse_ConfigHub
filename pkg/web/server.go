package web

import (
	"ConfigHub/pkg/datastore"
	"ConfigHub/pkg/logger"
	"ConfigHub/pkg/tools"
	"ConfigHub/pkg/web/handlers"
	"context"
	"crypto/tls"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	// "text/template"
    "html/template"
)

//go:embed templates/* static/css/* static/js/* static/images/*
var tmplFS embed.FS

var (
    indexTemplate *template.Template
    loginTemplate *template.Template
    configGeneralTemplate *template.Template
    configSegmentTemplate *template.Template
)

func init() {
    staticFS, err := fs.Sub(tmplFS, "static")
    if err != nil {
        logger.Logger.Panicf("Error loading static files: %v", err)
    }

    indexTemplate, err = template.ParseFS(tmplFS, "templates/index.html", "templates/base.html")
    if err != nil {
        logger.Logger.Panicf("Error parsing index template: %v", err)
    }

    loginTemplate, err = template.ParseFS(tmplFS, "templates/login.html", "templates/base.html")
    if err != nil {
        logger.Logger.Panicf("Error parsing login template: %v", err)
    }

    configGeneralTemplate, err = template.ParseFS(tmplFS, "templates/config-general.html", "templates/base.html")
    if err != nil {
        logger.Logger.Panicf("Error parsing login template: %v", err)
    }

    configSegmentTemplate, err = template.ParseFS(tmplFS, "templates/config-segment.html", "templates/base.html")
    if err != nil {
        logger.Logger.Panicf("Error parsing login template: %v", err)
    }

    http.Handle("/", handlers.AuthMiddleware(handlers.HomeHandler(indexTemplate)))
    http.HandleFunc("/login", handlers.LoginHandler(loginTemplate))
    http.HandleFunc("/config/general", handlers.AuthMiddleware(handlers.ConfigGeneralHandler(configGeneralTemplate)))
    http.HandleFunc("/update-general-config", handlers.AuthMiddleware(handlers.UpdateGeneralConfigHandler))
    http.HandleFunc("/update-log-config", handlers.AuthMiddleware(handlers.UpdateLogConfigHandler))
    http.HandleFunc("/update-audit-config", handlers.AuthMiddleware(handlers.UpdateAuditConfigHandler))
    http.HandleFunc("/update-webserver-config", handlers.AuthMiddleware(handlers.UpdateWebServerConfigHandler))
    http.HandleFunc("/config/segment", handlers.AuthMiddleware(handlers.ConfigSegmentHandler(configSegmentTemplate)))
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
}

func Webserver() {
	done := make(chan bool)
    go startServer(done)
	<-done
}


func startServer(done chan bool) {
    var (
        server *http.Server
        err error
    )
	conf := datastore.GetConfig().WebServer

    if !tools.CheckPortAvailability(conf.Port) {
		logger.Logger.Errorf("Port is already in use. Cannot start the web server. Port: %d\n", conf.Port)
        return
    }

    server = &http.Server{
        Addr: ":" + strconv.Itoa(conf.Port),
        Handler: nil,
        TLSConfig: &tls.Config{},
    }
    go func() {
        <-done
        if err := server.Shutdown(context.Background()); err != nil {
            fmt.Println("Server Shutdown:", err)
			logger.Logger.Infof("Server Shutdown: %s\n", err)
        }
    }()

    fmt.Println("Server starting on port", conf.Port)
    logger.Logger.Infof("Server starting on port %d", conf.Port)
    err = server.ListenAndServeTLS(conf.SSLCertPath, conf.SSLKeyPath)
    if err != http.ErrServerClosed {
        fmt.Println("Server failed:", err)
        logger.Logger.Infof("Server failed: %s", err)
    }
}