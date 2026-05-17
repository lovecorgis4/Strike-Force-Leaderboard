package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

type AppData struct {
	Scores map[string]int `json:"scores"`
	Logs   []string       `json:"logs"`
}

var (
	data     AppData
	mutex    sync.Mutex
	filename = "leaderboard.json"
)

func saveFile() {
	bytes, _ := json.Marshal(data)
	os.WriteFile(filename, bytes, 0644)
}

func loadFile() {
	bytes, err := os.ReadFile(filename)
	if err == nil {
		json.Unmarshal(bytes, &data)
	}
	if data.Scores == nil {
		data.Scores = make(map[string]int)
	}
	if data.Logs == nil {
		data.Logs = []string{}
	}
}

func logActivity(message string) {
	localZone := time.FixedZone("CEST", 2*60*60)

	currentTime := time.Now().In(localZone).Format("3:04 PM")
	logEntry := fmt.Sprintf("%s (%s)", message, currentTime)

	data.Logs = append([]string{logEntry}, data.Logs...)
	if len(data.Logs) > 10 {
		data.Logs = data.Logs[:10]
	}
}
func main() {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("leaderboard.json not found, creating a new one...")
		initialData := []byte(`{"scores":{},"logs":[]}`)
		os.WriteFile(filename, initialData, 0644)
	}

	loadFile()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		keys := make([]string, 0, len(data.Scores))
		for name := range data.Scores {
			keys = append(keys, name)
		}
		sort.Slice(keys, func(i, j int) bool {
			return data.Scores[keys[i]] > data.Scores[keys[j]]
		})

		fmt.Fprint(w, `
			<html>
			<head>
				<style>
					@keyframes pulse {
						0% { box-shadow: 0 0 0 0 rgba(212, 175, 55, 0.7); }
						70% { box-shadow: 0 0 0 15px rgba(212, 175, 55, 0); }
						100% { box-shadow: 0 0 0 0 rgba(212, 175, 55, 0); }
					}
					button:hover {
						filter: brightness(1.2);
						transform: scale(1.1);
						transition: 0.1s;
					}
				</style>
			</head>
			<body style='font-family: sans-serif; background: #000000; color: white; text-align: center;'>
				<h1>Strike Force Leaderboard</h1>
				<table style='margin: auto; border: 1px solid #ffffff; padding: 20px; border-radius: 10px; background: #4b4a4a; min-width: 300px;'>
		`)

		for i, name := range keys {
			wins := data.Scores[name]
			rowStyle := "background: #767676;"

			if i == 0 {
				rowStyle = "background: linear-gradient(90deg, #d4af37, #f1c40f); color: black; font-weight: bold; animation: pulse 2s infinite;"
			}

			fmt.Fprintf(w, "<tr style='%s'>", rowStyle)
			fmt.Fprintf(w, "<td style='padding: 15px; text-align: left; border-bottom: 1px solid #444;'>#%d %s</td>", i+1, name)
			fmt.Fprintf(w, "<td style='padding: 15px; border-bottom: 1px solid #444;'><strong>%d Wins</strong></td>", wins)
			fmt.Fprintf(w, "<td style='padding: 15px; border-bottom: 1px solid #444;'><button style='cursor: pointer; padding: 8px 12px; background: #00ff22be; color: white; border: none; border-radius: 4px;' onclick=\"window.location.href='/win?name=%s'\">Add Win</button></td>", name)
			fmt.Fprintf(w, "<td style='padding: 15px; border-bottom: 1px solid #444;'><button style='cursor: pointer; padding: 8px 12px; background: #ff0000; color: white; border: none; border-radius: 4px;' onclick=\"window.location.href='/delete?name=%s'\">Delete</button></td>", name)
			fmt.Fprintf(w, "<td style='padding: 15px; border-bottom: 1px solid #444;'><button style='cursor: pointer; padding: 8px 12px; background: #00d5ff; color: white; border: none; border-radius: 4px;' onclick=\"window.location.href='/deletewin?name=%s'\">Delete win</button></td>", name)
			fmt.Fprint(w, "</tr>")
		}

		fmt.Fprint(w, "</table>")

		fmt.Fprint(w, `
			<div style='margin-top: 30px; border-top: 1px solid #444; padding-top: 20px;'>
				<h3>Manage Players</h3>
				<input type='text' id='playerName' placeholder='Enter name' style='padding: 8px;'>
				<button style='cursor: pointer; padding: 8px 12px; background: #0008ff; color: white; border: none; border-radius: 4px;' 
					onclick="let name = document.getElementById('playerName').value; if(name) window.location.href='/add?name=' + name">
					Add New Player
				</button>
				<br><br>
				<button style='cursor: pointer; padding: 8px 12px; background: #7a0000; color: white; border: none; border-radius: 4px; font-size: 12px;' 
					onclick="if(confirm('Are you absolutely sure you want to WIPE the entire leaderboard? This cannot be undone.')) window.location.href='/reset'">
					⚠️ Reset Leaderboard
				</button>
			</div>
		`)

		fmt.Fprint(w, "<div style='margin-top: 40px; margin-bottom: 40px; border-top: 1px solid #444; padding-top: 20px;'>")
		fmt.Fprint(w, "<h3>Recent Activity Log</h3>")
		fmt.Fprint(w, "<ul style='list-style-type: none; padding: 0; max-width: 400px; margin: auto;'>")
		for _, logMsg := range data.Logs {
			fmt.Fprintf(w, "<li style='background: #222; margin: 5px 0; padding: 10px; border-radius: 4px; border-left: 4px solid #00ff22be; text-align: left;'>%s</li>", logMsg)
		}
		if len(data.Logs) == 0 {
			fmt.Fprint(w, "<li style='color: #888;'>No activity logged yet.</li>")
		}
		fmt.Fprint(w, "</ul></div>")

		fmt.Fprint(w, "</body></html>")
	})

	http.HandleFunc("/win", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name != "" {
			mutex.Lock()
			data.Scores[name]++
			logActivity(fmt.Sprintf("🏆 %s won a match!", name))
			saveFile()
			mutex.Unlock()
		}
		fmt.Fprint(w, "<html><body><script>window.location.href='/';</script></body></html>")
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name != "" {
			mutex.Lock()
			if _, exists := data.Scores[name]; !exists {
				data.Scores[name] = 0
				logActivity(fmt.Sprintf("👤 Added new player: %s", name))
				saveFile()
			}
			mutex.Unlock()
		}
		fmt.Fprint(w, "<html><body><script>window.location.href='/';</script></body></html>")
	})

	http.HandleFunc("/deletewin", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name != "" {
			mutex.Lock()
			if currentWins, exists := data.Scores[name]; exists && currentWins > 0 {
				data.Scores[name]--
				logActivity(fmt.Sprintf("📉 Removed a win from %s", name))
				saveFile()
			}
			mutex.Unlock()
		}
		fmt.Fprint(w, "<html><body><script>window.location.href='/';</script></body></html>")
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name != "" {
			mutex.Lock()
			if _, exists := data.Scores[name]; exists {
				delete(data.Scores, name)
				logActivity(fmt.Sprintf("❌ Completely removed player: %s", name))
				saveFile()
			}
			mutex.Unlock()
		}
		fmt.Fprint(w, "<html><body><script>window.location.href='/';</script></body></html>")
	})

	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		data.Scores = make(map[string]int)
		data.Logs = []string{}
		logActivity("⚙️ Leaderboard was completely reset!")
		saveFile()
		mutex.Unlock()
		fmt.Fprint(w, "<html><body><script>window.location.href='/';</script></body></html>")
	})

	fmt.Println("Server starting on localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
