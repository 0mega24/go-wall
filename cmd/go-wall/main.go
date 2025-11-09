package main

import (
	"fmt"
	"log"

	"github.com/0mega24/go-wall/internal/wallpaper"
)

func main() {
	path, err := wallpaper.GetCurrentWallpaper()
	if err != nil {
		log.Fatal("Error:", err)
	}
	fmt.Println("Current wallpaper:", path)
}
