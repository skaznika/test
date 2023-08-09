package main

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	StatusPending = 0
	StatusDone    = 1
	StatusError   = 2
)

type TranscriptionJob struct {
	gorm.Model
	Status       int            `json:"job_status"`
	ModelSize    string         `json:"model_size"`
	Task         string         `json:"task"`
	Language     string         `json:"language"`
	Device       string         `json:"device"`
	FileName     string         `json:"file_name"`
	SourceUrl    string         `json:"source_url"`
	ResultJson   datatypes.JSON `json:"result"`
	Translations []Translation  `gorm:"ForeignKey:TranscriptionJobID" json:"translations"`
}

type Translation struct {
	gorm.Model
	TranscriptionJobID uint           `json:"transcription_job_id"`
	SourceLanguage     string         `json:"source_language"`
	TargetLanguage     string         `json:"target_language"`
	ResultJson         datatypes.JSON `json:"result"`
	Status             int            `json:"translation_status"`
}

type TranscriptionResult struct {
	Language string  `json:"language"`
	Duration float64 `json:"duration"`
	Segments []struct {
		AvgLogprob       float64 `json:"avg_logprob"`
		CompressionRatio float64 `json:"compression_ratio"`
		End              float64 `json:"end"`
		ID               int     `json:"id"`
		NoSpeechProb     float64 `json:"no_speech_prob"`
		Seek             float64 `json:"seek"`
		Start            float64 `json:"start"`
		Temperature      float64 `json:"temperature"`
		Text             string  `json:"text"`
		Tokens           []int   `json:"tokens"`
		Words            []Word  `json:"words"`
	} `json:"segments"`
	Text string `json:"text"`
}

type Word struct {
	End         float64 `json:"end"`
	Start       float64 `json:"start"`
	Word        string  `json:"word"`
	Probability float64 `json:"probability"`
}
