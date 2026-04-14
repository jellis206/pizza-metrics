package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type GrafanaAlert struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type GrafanaWebhook struct {
	Alerts []GrafanaAlert `json:"alerts"`
}

func main() {
	topic := os.Getenv("NTFY_TOPIC")
	if topic == "" {
		log.Fatal("NTFY_TOPIC environment variable is required")
	}
	ntfyURL := os.Getenv("NTFY_URL")
	if ntfyURL == "" {
		ntfyURL = "https://ntfy.sh"
	}
	grafanaURL := os.Getenv("GRAFANA_URL")
	if grafanaURL == "" {
		grafanaURL = "https://metrics.urjellis.com"
	}

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload GrafanaWebhook
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		for _, alert := range payload.Alerts {
			name := alert.Labels["alertname"]
			if name == "" {
				name = "Grafana Alert"
			}
			status := strings.ToUpper(alert.Status)
			title := fmt.Sprintf("[%s] %s", status, name)

			body := alert.Annotations["summary"]
			if body == "" {
				body = alert.Annotations["description"]
			}
			if body == "" {
				body = fmt.Sprintf("Alert %s is %s", name, alert.Status)
			}

			priority := "default"
			switch alert.Labels["severity"] {
			case "critical":
				priority = "urgent"
			case "warning":
				priority = "high"
			}

			tags := "warning"
			if alert.Labels["severity"] == "critical" {
				tags = "rotating_light"
			}
			if alert.Status == "resolved" {
				tags = "white_check_mark"
			}

			url := fmt.Sprintf("%s/%s", ntfyURL, topic)
			req, err := http.NewRequest("POST", url, strings.NewReader(body))
			if err != nil {
				log.Printf("error creating request: %v", err)
				continue
			}
			req.Header.Set("Title", title)
			req.Header.Set("Priority", priority)
			req.Header.Set("Tags", tags)

			dashboardURL := fmt.Sprintf("%s/d/cfgcyvspjtq0wa/pizza-dashboard", grafanaURL)
			logsURL := fmt.Sprintf("%s/d/cfgcyvspjtq0wa/pizza-dashboard?viewPanel=11", grafanaURL)
			actions := fmt.Sprintf("view, Dashboard, %s; view, Logs, %s", dashboardURL, logsURL)
			req.Header.Set("Actions", actions)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("error sending to ntfy: %v", err)
				continue
			}
			resp.Body.Close()
			log.Printf("sent alert %q (%s) to ntfy, status=%d", name, alert.Status, resp.StatusCode)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	log.Println("ntfy-bridge listening on :9199")
	log.Fatal(http.ListenAndServe(":9199", nil))
}
