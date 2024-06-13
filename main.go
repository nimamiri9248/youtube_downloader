package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/Jeffail/gabs/v2"
)

type VideoInfo struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func main() {
	videoURL := "https://youtu.be/tSjnf6l8cq8?si=z4mD_v6sKvCY7ift" // Replace with your desired video URL

	videoInfo, err := getVideoInfo(videoURL)
	if err != nil {
		log.Fatalf("Failed to get video info: %v", err)
	}

	fmt.Printf("Downloading %s...\n", videoInfo.Title)
	err = downloadVideo(videoInfo.Url, videoInfo.Title+".mp4")
	if err != nil {
		log.Fatalf("Failed to download video: %v", err)
	}

	fmt.Println("Video downloaded successfully.")
}

func getVideoInfo(videoURL string) (*VideoInfo, error) {
	client := &http.Client{
		Timeout: 10 * time.Second, // Set a timeout for the HTTP request
	}

	resp, err := client.Get(videoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`ytInitialPlayerResponse\s*=\s*({.+?});`)
	match := re.FindSubmatch(body)
	if match == nil {
		return nil, fmt.Errorf("unable to find ytInitialPlayerResponse in HTML")
	}

	jsonStr := match[1]

	parsedJSON, err := gabs.ParseJSON(jsonStr)
	if err != nil {
		return nil, err
	}

	title := parsedJSON.Path("videoDetails.title").Data().(string)
	streamingData := parsedJSON.Path("streamingData").String()

	var streams map[string]interface{}
	json.Unmarshal([]byte(streamingData), &streams)

	formats := streams["formats"].([]interface{})
	if len(formats) == 0 {
		return nil, fmt.Errorf("no video formats found")
	}

	bestFormat := formats[0].(map[string]interface{})
	videoURL = bestFormat["url"].(string)

	return &VideoInfo{
		Title: title,
		Url:   videoURL,
	}, nil
}

func downloadVideo(videoURL string, filename string) error {
	client := &http.Client{
		Timeout: 0, // No timeout for downloading
	}

	resp, err := client.Get(videoURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download video, status code: %d", resp.StatusCode)
	}

	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
