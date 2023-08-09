package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func formatTime(t float64) string {
	ms := int(t * 1000)
	s := ms / 1000
	m := s / 60
	h := m / 60
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m%60, s%60, ms%1000)
}

func JsonToSrt(j *TranscriptionJob, r *TranscriptionResult) (string, error) {
	o := fmt.Sprintf("%s/%v-%s_%s.srt", UPLOAD_DIR, j.ID, j.FileName, r.Language)
	f, err := os.Create(o)
	if err != nil {
		Lggr.Error().Err(err).Msg("Error creating file")
		return "", err
	}
	defer f.Close()

	Lggr.Printf("Segments %v", r.Segments)
	for n, s := range r.Segments {
		start := s.Start
		end := s.End
		text := s.Text
		fmt.Fprintf(f, "%d\n%s --> %s\n%s\n\n", n+1, formatTime(start), formatTime(end), strings.TrimSpace(text))
	}

	return o, nil
}

func YtDlp(url string, id int64) (string, error) {
	formatOutput := func(title string) string {
		return fmt.Sprintf("%s/%v-%s.%s", UPLOAD_DIR, id, title, "%(ext)s")
	}

	// Get video ID and title
	getIDTitleCmd := exec.Command("yt-dlp", "--get-id", "--get-title", url)
	output, err := getIDTitleCmd.CombinedOutput()
	if err != nil {
		Lggr.Error().Err(err).Msg("Error getting video ID and title")
		return "", err
	}

	// Process output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	titleRegex := regexp.MustCompile("[^a-zA-Z0-9]+")
	cleanTitle := titleRegex.ReplaceAllString(lines[0], "_")
	downloadCmd := exec.Command("yt-dlp", "-o", formatOutput(cleanTitle), url)

	// Download video
	if err = downloadCmd.Run(); err != nil {
		Lggr.Error().Err(err).Msg("Error downloading video")
		return "", err
	}

	// Find downloaded file
	files, _ := ioutil.ReadDir(UPLOAD_DIR)
	for _, file := range files {
		prefix := fmt.Sprintf("%v-%s.", id, cleanTitle)
		if strings.HasPrefix(file.Name(), prefix) {
			ext := strings.Split(file.Name(), ".")[1]
			return cleanTitle + "." + ext, nil
		}
	}

	return "", fmt.Errorf("file not found")
}

func TranslateString(text string, sourceLang string, targetLang string) (string, error) {
	Lggr.Printf("Translating %v from %v to %v", text, sourceLang, targetLang)

	client := &http.Client{}
	data := map[string]string{"q": text, "source": sourceLang, "target": targetLang}
	jsonData, _ := json.Marshal(data)
	Lggr.Printf("Sending %v", string(jsonData))

	req, _ := http.NewRequest("POST", fmt.Sprintf("http://%v/translate", LT_ENDPOINT), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		Lggr.Error().Err(err).Msg("Error sending request")
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	translated := result["translatedText"].(string)

	Lggr.Printf("Translated %v to %v", text, translated)

	return translated, nil
}
