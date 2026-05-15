# Strike Force Leaderboard

A lightweight, web-based leaderboard system built with **Go** and **Docker**. This project allows users to track wins, add new players, and manage rankings in real-time.

## Features
* **Live Leaderboard:** Automatically sorts players by win count.
* **CRUD Operations:** Add players, record wins, and remove players/wins.
* **Visual Styling:** The #1 player is highlighted with a gold gradient.
* **Persistent Storage:** Data is saved to `leaderboard.json` so scores aren't lost when the app closes.

## Getting Started

### Prerequisites
You will need to have [Docker](https://www.docker.com/) installed on your machine.

### Installation & Setup

 **Clone the repository:**
   To install this leaderboard you need to open a terminal in the folder you wnt the code and files in, run this command git clone https://github.com/vsc8080/Strike-Force-Leaderboard.git. Once you have the files run these commands:
  ' docker build -t strike-force-leaderboard . ' (to build the docker image)

  ' docker run -p 8080:8080 -v ${PWD}/leaderboard.json:/root/leaderboard.json leaderboard-app ' for Windows users
  ' docker run -p 8080:8080 -v $(pwd)/leaderboard.json:/root/leaderboard.json leaderboard-app ' for Mac/Linux users 
  (this is what runs the program)
After this you go to localhost:8080 and there you have the leaderboard (if you want to stop the current container, just press Ctrl + C)