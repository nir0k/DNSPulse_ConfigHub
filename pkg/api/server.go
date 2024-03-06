package api

import (
	"DNSPulse_ConfigHub/pkg/datastore"
	"DNSPulse_ConfigHub/pkg/logger"
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed html/*.html static/*
var tmplFS embed.FS

func Apisrv() {

    staticFS, err := fs.Sub(tmplFS, "static")
    if err != nil {
        panic(err)
    }
	conf := datastore.GetConfig().Api
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/login", loginHandler)
    router.StaticFS("/static", http.FS(staticFS))
	configRoutes := router.Group("/api/config/general")
	configRoutes.Use(tokenAuthMiddleware())
	{
		configRoutes.GET("/main", getMainConfig)
		configRoutes.PATCH("/main", updateMainConfig)
		configRoutes.PUT("/main", updateMainConfig)

		configRoutes.GET("/log", getLogConfig)
		configRoutes.PATCH("/log", updateLogConfig)
		configRoutes.PUT("/log", updateLogConfig)

		configRoutes.GET("/audit", getAuditConfig)
		configRoutes.PATCH("/audit", updateAuditConfig)
		configRoutes.PUT("/audit", updateAuditConfig)

		configRoutes.GET("/web", getWebConfig)
		configRoutes.PATCH("/web", updateWebConfig)
		configRoutes.PUT("/web", updateWebConfig)

		configRoutes.GET("/api", getApiConfig)
		configRoutes.PATCH("/api", updateApiConfig)
		configRoutes.PUT("/api", updateApiConfig)

		configRoutes.GET("/segment", getSegmentConfig)
		configRoutes.POST("/segment", createSegmentConfig)
		configRoutes.GET("/segment/:name", getSpecificSegmentConfig)
		configRoutes.PATCH("/segment/:name", updateSpecificSegmentConfig)
		configRoutes.PUT("/segment/:name", updateSpecificSegmentConfig)
    	configRoutes.DELETE("/segment/:name", deleteSegmentConfig)
		
		configRoutes.GET("/file", DownloadConfigHandler)
        configRoutes.POST("/file", UploadConfigHandler)
	}

	configRoutes = router.Group("/api/config/segment")
	configRoutes.Use(tokenAuthMiddleware())
	{
		configRoutes.GET("/:name/main", getSegmentMainConfig)
		configRoutes.PATCH("/:name/main", updateSegmentMainConfig)
		configRoutes.PUT("/:name/main", updateSegmentMainConfig)

		configRoutes.GET("/:name/sync", getSegmentSyncConfig)
		configRoutes.PATCH("/:name/sync", updateSegmentSyncConfig)
		configRoutes.PUT("/:name/sync", updateSegmentSyncConfig)

		configRoutes.GET("/:name/polling", getSegmentPollingConfig)
		configRoutes.PATCH("/:name/polling", updateSegmentPollingConfig)
		configRoutes.PUT("/:name/polling", updateSegmentPollingConfig)

		configRoutes.GET("/:name/prometheus", getSegmentPrometheusConfig)
		configRoutes.PATCH("/:name/prometheus", updateSegmentPrometheusConfig)
		configRoutes.PUT("/:name/prometheus", updateSegmentPrometheusConfig)

		configRoutes.GET("/:name/prometheus/label", getSegmentPrometheusLabelsConfig)
		configRoutes.PATCH("/:name/prometheus/label", updateSegmentPrometheusLabelConfig)
		configRoutes.PUT("/:name/prometheus/label", updateSegmentPrometheusLabelConfig)

		configRoutes.GET("/:name/polling/servers", getSegmentPollingHostsConfig)
		configRoutes.POST("/:name/polling/servers", addPollingServerConfig)
		configRoutes.GET("/:name/polling/servers/:srvname", getPollingServerConfig)
		configRoutes.PATCH("/:name/polling/servers/:srvname", updatePollingServerConfig)
		configRoutes.PUT("/:name/polling/servers/:srvname", updatePollingServerConfig)
		configRoutes.DELETE("/:name/polling/servers/:srvname", deletePollingServerConfig)

		configRoutes.GET("/:name/file", DownloadSegmentConfigHandler)
        configRoutes.POST("/:name/file", UploadSegmentConfigHandler)

		configRoutes.GET("/:name/polling/servers/file", DownloadSegmentPollingHandler)
        configRoutes.POST("/:name/polling/servers/file", UploadSegmentPollingHandler)
	}

    router.GET("/docs", wrapHandler(ServeDocs(tmplFS)))
    router.GET("/api/api-spec", wrapHandler(ServeAPISpec(staticFS, conf.Port)))
	
	fmt.Printf("API starting on port %d. TLS: %v\n", conf.Port, conf.SSLEnabled)
    logger.Logger.Infof("Web Server starting on %d. TLS: %v", conf.Port, conf.SSLEnabled)
	if conf.SSLEnabled {
		err = router.RunTLS(fmt.Sprintf(":%d", conf.Port), conf.SSLCertPath, conf.SSLKeyPath)
	} else {
		err = router.Run(fmt.Sprintf(":%d", conf.Port))
	}
	
	if err != nil {
		fmt.Printf("Failed to start API on port%d. TLS: %v\n", conf.Port, conf.SSLEnabled)
		logger.Logger.Fatalf("Failed to start server: %v", err)
	}
	
}

func wrapHandler(h func(http.ResponseWriter, *http.Request)) gin.HandlerFunc {
    return func(c *gin.Context) {
        h(c.Writer, c.Request)
    }
}
