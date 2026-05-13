package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ServicesHandler(services []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"emulator": "gcs-emulator",
			"version":  "0.1.0",
			"services": services,
		})
	}
}

func DashboardHandler(services []string, stats func() map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/-/" {
			http.NotFound(w, r)
			return
		}
		s := stats()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>GCP Emulator</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,system-ui,sans-serif;background:#1a1a2e;color:#e0e0e0;min-height:100vh}
header{background:#16213e;padding:20px 30px;display:flex;justify-content:space-between;align-items:center;border-bottom:1px solid #0f3460}
h1{font-size:20px;color:#e94560}
.version{font-size:12px;color:#888}
main{display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:20px;padding:30px;max-width:1200px;margin:0 auto}
.card{background:#16213e;border-radius:8px;padding:20px;border:1px solid #0f3460}
.card h2{font-size:14px;color:#e94560;margin-bottom:12px;text-transform:uppercase;letter-spacing:1px}
.service{display:flex;justify-content:space-between;align-items:center;padding:8px 0;border-bottom:1px solid #0f3460}
.service:last-child{border-bottom:none}
.service .name{font-size:14px}
.service .status{font-size:11px;padding:2px 8px;border-radius:10px}
.status-up{background:#1b5e20;color:#a5d6a7}
.stat{margin:6px 0;font-size:13px;display:flex;justify-content:space-between}
.stat .label{color:#aaa}
.stat .value{color:#e94560;font-weight:bold;font-family:monospace}
.metrics{grid-column:1/-1}
.metrics pre{font-size:11px;color:#888;max-height:200px;overflow-y:auto;font-family:monospace}
footer{text-align:center;padding:20px;color:#555;font-size:11px}
</style>
</head>
<body>
<header><h1>GCP Emulator</h1><span class="version">v0.1.0</span></header>
<main>
<div class="card">
<h2>Services</h2>
`)
		for _, svc := range services {
			fmt.Fprintf(w, `<div class="service"><span class="name">%s</span><span class="status status-up">running</span></div>`, svc)
		}

		fmt.Fprint(w, `</div>
<div class="card">
<h2>Stats</h2>
`)
		if s != nil {
			for k, v := range s {
				fmt.Fprintf(w, `<div class="stat"><span class="label">%s</span><span class="value">%v</span></div>`, k, v)
			}
		}
		fmt.Fprint(w, `</div>`)

		fmt.Fprint(w, `</main><footer>GCP Emulator · port 9090 · no credentials required</footer></body></html>`)
	}
}
