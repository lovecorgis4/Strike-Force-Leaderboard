package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
)

var (
	leaderboard = make(map[string]int)
	mutex       sync.Mutex
	filename    = "leaderboard.json"
)

func saveFile() {
	data, _ := json.Marshal(leaderboard)
	os.WriteFile(filename, data, 0644)
}

func loadFile() {
	data, err := os.ReadFile(filename)
	if err == nil {
		json.Unmarshal(data, &leaderboard)
	}
}

func main() {
	// 1. Auto-create file if missing
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("leaderboard.json not found, creating a new one...")
		initialData := []byte("{}")
		os.WriteFile(filename, initialData, 0644)
	}

	// 2. Load the data into memory
	loadFile()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// 3. Logic: Prepare and Sort the keys
		keys := make([]string, 0, len(leaderboard))
		for name := range leaderboard {
			keys = append(keys, name)
		}
		sort.Slice(keys, func(i, j int) bool {
			return leaderboard[keys[i]] > leaderboard[keys[j]]
		})

		// 4. Start HTML and Style Tags
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

		// 5. Build Table Rows
		for i, name := range keys {
			wins := leaderboard[name]
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

		// 6. Manage Players Section
		fmt.Fprint(w, `
			<div style='margin-top: 30px; border-top: 1px solid #444; padding-top: 20px;'>
				<h3>Manage Players</h3>
				<input type='text' id='playerName' placeholder='Enter name' style='padding: 8px;'>
				<button style='cursor: pointer; padding: 8px 12px; background: #0008ff; color: white; border: none; border-radius: 4px;' 
					onclick="let name = document.getElementById('playerName').value; if(name) window.location.href='/add?name=' + name">
					Add New Player
				</button>
			</div>
			</body></html>
		`)
	})

	// --- ROUTES ---

	http.HandleFunc("/win", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name != "" {
			mutex.Lock()
			leaderboard[name]++
			saveFile()
			mutex.Unlock()
		}
		fmt.Fprint(w, "<html><body><script>window.location.href='/';</script></body></html>")
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name != "" {
			mutex.Lock()
			if _, exists := leaderboard[name]; !exists {
				leaderboard[name] = 0
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
			if currentWins, exists := leaderboard[name]; exists && currentWins > 0 {
				leaderboard[name]--
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
			delete(leaderboard, name)
			saveFile()
			mutex.Unlock()
		}
		fmt.Fprint(w, "<html><body><script>window.location.href='/';</script></body></html>")
	})

	fmt.Println("Server starting on localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
