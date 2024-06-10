package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var wall = '▤'
var character = '☺'

func generateMaze(width int, height int) [][]rune {
	maze := make([][]rune, height)
	for i := range maze {
		maze[i] = make([]rune, width)
		for j := range maze[i] {
			maze[i][j] = wall
		}
	}

	rand.Seed(time.Now().UnixNano())

	for i := 1; i < height-1; i += 2 {
		for j := 1; j < width-1; j += 2 {
			maze[i][j] = ' '

			if j < width-2 {
				if rand.Intn(2) == 0 {
					maze[i][j+1] = ' '
				} else {
					maze[i][j+1] = wall
				}
			}
			if i < height-2 {
				if rand.Intn(2) == 0 {
					maze[i+1][j] = ' '
				} else {
					maze[i+1][j] = wall
				}
			}
		}
	}

	// Escolha aleatoriamente uma posição inicial para o personagem em um espaço vazio
	for {
		startX := rand.Intn(width)
		startY := rand.Intn(height)
		if maze[startY][startX] == ' ' {
			maze[startY][startX] = character
			break // Sai do loop uma vez que uma posição válida foi encontrada
		}
	}

	return maze
}

func printMaze(maze [][]rune) {
	for _, row := range maze {
		for _, c := range row {
			fmt.Print(string(c))
		}
		fmt.Println()
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run maze.go <width> <height>")
		return
	}
	width, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid width")
		return
	}
	height, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Invalid height")
		return
	}

	maze := generateMaze(width, height)
	printMaze(maze)
}
