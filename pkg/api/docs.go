package api

import (
	"DNSPulse_ConfigHub/pkg/logger"
	"DNSPulse_ConfigHub/pkg/tools"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

func ServeAPISpec(staticFS fs.FS, port int) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        specFile := "api-spec.yaml"
    
        fileData, err := fs.ReadFile(staticFS, specFile)
        if err != nil {
            http.Error(w, "File not found", http.StatusNotFound)
            return
        }
        specContent := string(fileData)
        
        serverURL := fmt.Sprintf("https://%s:%d", tools.GetHostName(), port)
        specContent = strings.Replace(specContent, "SERVER_URL_PLACEHOLDER", serverURL, -1)

        w.Header().Set("Content-Type", "application/yaml")
    
        w.Write([]byte(specContent))
    }
}

func ServeDocs(tmplFS fs.FS) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        htmlFile, err := fs.ReadFile(tmplFS, "html/redoc.html")
        if err != nil {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            logger.Logger.Errorf("Error reading redoc.html: %s", err)
            return
        }

        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Write(htmlFile)
    }
}
