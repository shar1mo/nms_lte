package httpapi

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>NMS LTE API Docs</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: '/swagger/doc.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis],
        layout: 'BaseLayout'
      });
    };
  </script>
</body>
</html>
`

func registerSwaggerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/swagger/", serveSwaggerUI)
	mux.HandleFunc("/swagger/doc.json", serveSwaggerJSON)
	mux.HandleFunc("/swagger/doc.yaml", serveSwaggerYAML)
}

func serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/swagger/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerUIHTML))
}

func serveSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	serveSwaggerFile(w, "swagger.json", "application/json; charset=utf-8")
}

func serveSwaggerYAML(w http.ResponseWriter, r *http.Request) {
	serveSwaggerFile(w, "swagger.yaml", "application/yaml; charset=utf-8")
}

func serveSwaggerFile(w http.ResponseWriter, name, contentType string) {
	path := filepath.Join("docs", "swagger", name)
	content, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("swagger file %q not found, run `make swagger` first", path), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(content)
}
