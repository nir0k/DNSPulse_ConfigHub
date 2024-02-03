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
    "html/template"
)

//go:embed templates/* static/css/* static/js/* static/images/* static/webfonts/*
var tmplFS embed.FS

var (
    indexTemplate *template.Template
    loginTemplate *template.Template
    configGeneralTemplate *template.Template
    configSegmentTemplate *template.Template
    pollingSegmentTemplate *template.Template
)

func setup() error {
    staticFS, err := fs.Sub(tmplFS, "static")
    if err != nil {
        return fmt.Errorf("failed loading static files: %v", err)
    }

    indexTemplate, err = template.ParseFS(tmplFS, "templates/index.html", "templates/base.html")
    if err != nil {
        return fmt.Errorf("failed parsing index template: %v", err)
    }

    loginTemplate, err = template.ParseFS(tmplFS, "templates/login.html", "templates/base.html")
    if err != nil {
        return fmt.Errorf("failed parsing login template: %v", err)
    }

    configGeneralTemplate, err = template.ParseFS(tmplFS, "templates/config-general.html", "templates/base.html")
    if err != nil {
        return fmt.Errorf("failed parsing config-general template: %v", err)
    }

    configSegmentTemplate, err = template.ParseFS(tmplFS, "templates/config-segment.html", "templates/base.html")
    if err != nil {
        return fmt.Errorf("failed parsing config-segment template: %v", err)
    }

    pollingSegmentTemplate, err = template.ParseFS(tmplFS, "templates/config-polling.html", "templates/base.html")
    if err != nil {
        return fmt.Errorf("failed parsing config-polling template: %v", err)
    }

    http.Handle("/", handlers.AuthMiddleware(handlers.HomeHandler(indexTemplate)))
    http.HandleFunc("/login", handlers.LoginHandler(loginTemplate))
    
    http.HandleFunc("/config/general", handlers.AuthMiddleware(handlers.ConfigGeneralHandler(configGeneralTemplate)))
    http.HandleFunc("/config/general/update-general", handlers.AuthMiddleware(handlers.UpdateGeneralConfigHandler))
    http.HandleFunc("/config/general/update-log", handlers.AuthMiddleware(handlers.UpdateLogConfigHandler))
    http.HandleFunc("/config/general/update-audit", handlers.AuthMiddleware(handlers.UpdateAuditConfigHandler))
    http.HandleFunc("/config/general/update-webserver", handlers.AuthMiddleware(handlers.UpdateWebServerConfigHandler))
    http.HandleFunc("/config/general/download", handlers.AuthMiddleware(handlers.DownloadConfigHandler))
    http.HandleFunc("/config/general/upload", handlers.AuthMiddleware(handlers.DownloadConfigHandler))
    
    http.HandleFunc("/config/segment", handlers.AuthMiddleware(handlers.ConfigSegmentHandler(configSegmentTemplate)))
    http.HandleFunc("/config/segment/update", handlers.AuthMiddleware(handlers.UpdateSegmentConfigHandler))
    http.HandleFunc("/config/segment/download", handlers.AuthMiddleware(handlers.DownloadSegmentConfigHandler))
    http.HandleFunc("/config/segment/upload", handlers.AuthMiddleware(handlers.UploadSegmentConfigHandler))

    http.HandleFunc("/config/polling", handlers.AuthMiddleware(handlers.PollingSegmentHandler(pollingSegmentTemplate)))
    http.HandleFunc("/config/polling/file", handlers.AuthMiddleware(handlers.FileHandler))
    http.HandleFunc("/config/polling/update", handlers.AuthMiddleware(handlers.UpdateSegmentDataHandler))

    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
    
    return nil
}

func Webserver() {
    err := setup()
    if err != nil {
        logger.Logger.Errorf("Web server setup failed: %v", err)
    }
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
    serverAddr := ":" + strconv.Itoa(conf.Port)
    if conf.ListenAddress != "0.0.0.0" {
        serverAddr = conf.ListenAddress + serverAddr
    }
    if conf.SSLEnabled {
        server = &http.Server{
            Addr: serverAddr,
            Handler: nil,
            TLSConfig: &tls.Config{},
        }
    } else {
        server = &http.Server{
            Addr: serverAddr,
            Handler: nil,
        }
    }
    go func() {
        <-done
        if err := server.Shutdown(context.Background()); err != nil {
            fmt.Println("Server Shutdown:", err)
			logger.Logger.Infof("Server Shutdown: %s\n", err)
        }
    }()

    fmt.Printf("Web Server starting on %s\n", serverAddr)
    logger.Logger.Infof("Web Server starting on %s", serverAddr)
    if conf.SSLEnabled {
        err = server.ListenAndServeTLS(conf.SSLCertPath, conf.SSLKeyPath)
    } else {
        err = server.ListenAndServe()
    }
    if err != http.ErrServerClosed {
        fmt.Println("Server failed:", err)
        logger.Logger.Infof("Server failed: %s", err)
    }
}