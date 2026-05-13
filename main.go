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
	loadFile()

	// MAIN DASHBOARD
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		keys := make([]string, 0, len(leaderboard))
		for name := range leaderboard {
			keys = append(keys, name)
		}
		sort.Slice(keys, func(i, j int) bool {
			return leaderboard[keys[i]] > leaderboard[keys[j]]
		})

		fmt.Fprint(w, "<html><body style='font-family: sans-serif; background: #1a1a1a; color: white; text-align: center;'>")
		fmt.Fprint(w, "<h1>Strike Force Leaderboard</h1>")
		fmt.Fprint(w, "<table style='margin: auto; border: 1px solid #444; padding: 20px; border-radius: 10px; background: #2a2a2a; min-width: 300px;'>")

		// 4. Create the table rows
		for i, name := range keys {
			wins := leaderboard[name]

			// Default style
			rowStyle := "background: #2a2a2a;"

			// GOLD style for #1
			if i == 0 {
				rowStyle = "background: linear-gradient(90deg, #d4af37, #f1c40f); color: black; font-weight: bold;"
			}

			// Apply the style to the TR (Table Row)
			fmt.Fprintf(w, "<tr style='%s'>", rowStyle)

			// Name Column
			fmt.Fprintf(w, "<td style='padding: 15px; text-align: left; border-bottom: 1px solid #444;'>#%d %s</td>", i+1, name)

			// Wins Column
			fmt.Fprintf(w, "<td style='padding: 15px; border-bottom: 1px solid #444;'><strong>%d Wins</strong></td>", wins)

			// Buttons Columns
			fmt.Fprintf(w, "<td style='padding: 15px; border-bottom: 1px solid #444;'><button style='cursor: pointer; padding: 8px 12px; background: #4CAF50; color: white; border: none; border-radius: 4px;' onclick=\"window.location.href='/win?name=%s'\">Add Win</button></td>", name)
			fmt.Fprintf(w, "<td style='padding: 15px; border-bottom: 1px solid #444;'><button style='cursor: pointer; padding: 8px 12px; background: #ff4444; color: white; border: none; border-radius: 4px;' onclick=\"window.location.href='/delete?name=%s'\">Delete</button></td>", name)

			fmt.Fprint(w, "</tr>")
		}

		fmt.Fprint(w, "</table>")

		// ADD NEW PLAYER SECTION (Outside the table)
		fmt.Fprint(w, `
			<div style='margin-top: 30px; border-top: 1px solid #444; padding-top: 20px;'>
				<h3>Manage Players</h3>
				<input type='text' id='playerName' placeholder='Enter name' style='padding: 8px;'>
				<button style='cursor: pointer; padding: 8px 12px; background: #0008ff; color: white; border: none; border-radius: 4px;' 
					onclick="let name = document.getElementById('playerName').value; if(name) window.location.href='/add?name=' + name">
					Add New Player
				</button>
			</div>
		`)

		fmt.Fprint(w, "</body></html>")
	})

	// ADD WIN ROUTE
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

	// ADD PLAYER ROUTE
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

	// DELETE PLAYER ROUTE
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name != "" {
			mutex.Lock()
			// Only delete if they ARE in the map
			if _, exists := leaderboard[name]; exists {
				delete(leaderboard, name)
				saveFile()
			}
			mutex.Unlock()
		}
		fmt.Fprint(w, "<html><body><script>window.location.href='/';</script></body></html>")
	})

	fmt.Println("Server starting on localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
