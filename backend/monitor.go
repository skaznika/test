package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"gorm.io/datatypes"
)

func MonitorJobs() {
	var transcriptionJobs []TranscriptionJob
	var translationJobs []Translation

	for {
		// Process transcription jobs
		DBCon.Where("status = ?", StatusPending).Find(&transcriptionJobs)
		for _, job := range transcriptionJobs {
			asrResponse, err := processTranscriptionJob(&job, ASR_ENDPOINT)
			if err == nil {
				updateTranscriptionJobSuccess(&job, asrResponse)
			} else {
				Lggr.Error().Err(err).Msg("Error processing job")
				updateTranscriptionJobError(&job)
			}
		}

		// Process translation jobs
		DBCon.Where("status = ?", StatusPending).Find(&translationJobs)
		for _, job := range translationJobs {
			Lggr.Printf("Processing translation job: %v", job.ID)
			_ = processTranslationJob(&job)
		}

		time.Sleep(3 * time.Second)
	}
}

func processTranscriptionJob(job *TranscriptionJob, ASR_ENDPOINT string) (TranscriptionResult, error) {
	var asrResponse TranscriptionResult
	body, writer, err := prepareMultipartFormData(job)
	if err != nil {
		Lggr.Error().Err(err).Msg("Error preparing multipart form data")
		return asrResponse, err
	}

	url := fmt.Sprintf("http://%v/transcribe?model_size=%v&task=%v&language=%v&device=%v", ASR_ENDPOINT, job.ModelSize, job.Task, job.Language, job.Device)
	Lggr.Printf("Sending request to %v", url)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		Lggr.Printf("Error creating request: %v", err)
		return asrResponse, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Lggr.Error().Err(err).Msg("Error sending request")
		return asrResponse, err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		Lggr.Error().Msgf("Response from %v: %v", url, string(b))
		Lggr.Error().Err(err).Msgf("Invalid response status %v:", resp.StatusCode)
		return asrResponse, errors.New("invalid status")
	}

	Lggr.Printf("Response: %v", string(b))
	if err := json.Unmarshal(b, &asrResponse); err != nil {
		Lggr.Error().Err(err).Msg("Error decoding response")
		return asrResponse, err
	}

	return asrResponse, nil
}

func prepareMultipartFormData(job *TranscriptionJob) (*bytes.Buffer, *multipart.Writer, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", job.FileName)
	if err != nil {
		Lggr.Error().Err(err).Msg("Error creating form file")
		return nil, nil, err
	}

	// Read file from disk
	filePath := fmt.Sprintf("%v/%v-%v", UPLOAD_DIR, job.ID, job.FileName)
	file, err := os.Open(filePath)
	if err != nil {
		Lggr.Error().Err(err).Msg("Error opening file")
		return nil, nil, err
	}
	defer file.Close()

	_, err = io.Copy(part, file)
	if err != nil {
		Lggr.Error().Err(err).Msg("Error copying file")
		return nil, nil, err
	}

	err = writer.Close()
	if err != nil {
		Lggr.Error().Err(err).Msg("Error closing writer")
		return nil, nil, err
	}

	return body, writer, nil
}

func updateTranscriptionJobSuccess(job *TranscriptionJob, asrResponse TranscriptionResult) {
	job.Status = StatusDone
	resJson, err := json.Marshal(asrResponse)
	if err != nil {
		Lggr.Error().Err(err).Msg("Error encoding response json")
		return
	}
	job.ResultJson = datatypes.JSON(resJson)
	err = DBCon.Save(job).Error
	if err != nil {
		Lggr.Error().Err(err).Msg("Error saving job")
	}
}

func updateTranscriptionJobError(job *TranscriptionJob) {
	job.Status = StatusError
	DBCon.Save(job)
}

func processTranslationJob(job *Translation) error {
	var result TranscriptionResult

	// Get the original result from the job
	var transcriptionJob TranscriptionJob
	err := DBCon.First(&transcriptionJob, job.TranscriptionJobID).Error
	if err != nil {
		Lggr.Error().Err(err).Msg("Error getting transcription job")
		return err
	}

	// Iterate the job result json
	err = json.Unmarshal(transcriptionJob.ResultJson, &result)
	if err != nil {
		job.Status = StatusError
		errt := DBCon.Save(job).Error
		if errt != nil {
			Lggr.Error().Err(errt).Msg("Error saving job")
		}
		Lggr.Error().Err(err).Msg("Error decoding result json")
		return err
	}

	Lggr.Printf("Processing translation job: %v", job)

	// Create a new map for the translated result
	translatedResult := result

	for i, segment := range result.Segments {
		// Translate the text
		translatedText, err := TranslateString(segment.Text, result.Language, job.TargetLanguage)
		if err != nil {
			job.Status = StatusError
			errt := DBCon.Save(job).Error
			if errt != nil {
				Lggr.Error().Err(errt).Msg("Error saving job")
			}
			Lggr.Error().Err(err).Msg("Error translating text")
			return err
		}
		// Update the translated text in the segment
		translatedResult.Segments[i].Text = translatedText
	}

	// Combine the translated segments into a single string
	translatedResult.Text = ""
	for _, segment := range translatedResult.Segments {
		translatedResult.Text += segment.Text + " "
	}

	// Remove any trailing space
	translatedResult.Text = strings.TrimSpace(translatedResult.Text)

	// Update the translated result
	translatedResult.Language = job.TargetLanguage

	translatedResultJson, err := json.Marshal(translatedResult)
	if err != nil {
		job.Status = StatusError
		errt := DBCon.Save(job).Error
		if errt != nil {
			Lggr.Error().Err(errt).Msg("Error saving job")
			return errt
		}
		Lggr.Error().Err(err).Msg("Error encoding translated result json")
		return err
	}
	job.ResultJson = datatypes.JSON(translatedResultJson)
	job.Status = StatusDone
	// Save the job to the database
	err = DBCon.Save(job).Error
	if err != nil {
		Lggr.Error().Err(err).Msg("Error saving job")
		return err
	}
	return nil
}
