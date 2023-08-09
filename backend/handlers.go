package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"gorm.io/datatypes"
)

func ProcessFrom(w http.ResponseWriter, r *http.Request) {
	Lggr.Printf("Form data: %v", r.Form)
	Lggr.Printf("Source: %v", r.FormValue("source"))
	if err := r.ParseForm(); err != nil {
		Lggr.Error().Err(err).Msg("Invalid data")
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	filename := ""
	// Get las job Id from database and increment
	var count int64
	DBCon.Unscoped().Model(&TranscriptionJob{}).Where("ID > ?", "0").Count(&count)
	newId := count + 1

	if r.FormValue("source") == "file" {
		file, header, err := r.FormFile("file")
		if err != nil {
			Lggr.Error().Err(err).Msg("Invalid file")
			http.Error(w, "Invalid file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		filename = header.Filename

		// Store file on disk
		Lggr.Printf("New job id: %v", newId)
		dst := fmt.Sprintf("%v/%v-%v", UPLOAD_DIR, newId, header.Filename)
		f, err := os.Create(dst)
		if err != nil {
			Lggr.Error().Err(err).Msg("Error saving file")
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		if _, err := io.Copy(f, file); err != nil {
			Lggr.Error().Err(err).Msg("Error saving file")
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
	}

	if r.FormValue("source") == "url" {
		downloadedFile, err := YtDlp(r.FormValue("url"), newId)
		filename = downloadedFile
		if err != nil {
			Lggr.Error().Err(err).Msg("Error downloading file")
			http.Error(w, "Error downloading file", http.StatusInternalServerError)
			return
		}
	}

	job := TranscriptionJob{
		Status:    StatusPending,
		ModelSize: r.FormValue("model_size"),
		Task:      r.FormValue("task"),
		Language:  r.FormValue("language"),
		FileName:  filename,
		Device:    "cpu",
	}

	if err := DBCon.Create(&job).Error; err != nil {
		Lggr.Error().Err(err).Msg("Error saving job")
		http.Error(w, "Error saving job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetJobs(w http.ResponseWriter, r *http.Request) {
	var jobs []TranscriptionJob
	DBCon.Preload("Translations", "status = ?", StatusDone).Order("created_at desc").Find(&jobs)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func GetJobById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	Lggr.Printf("Getting job %v", id)

	// Get job from database with GORM
	var job TranscriptionJob
	if err := DBCon.Preload("Translations", "status = ?", StatusDone).First(&job, id).Error; err != nil {
		Lggr.Error().Err(err).Msg("Error getting job")
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func GetSubs(w http.ResponseWriter, r *http.Request) {
	var job TranscriptionJob
	id := chi.URLParam(r, "id")
	format := chi.URLParam(r, "format")
	lang := r.URL.Query().Get("lang")

	Lggr.Print("Getting subs")
	Lggr.Printf("Lang %v", lang)
	Lggr.Printf("Id %v", id)
	Lggr.Printf("Format %v", format)

	if err := DBCon.Preload("Translations", "status = ?", StatusDone).First(&job, id).Error; err != nil {
		Lggr.Error().Err(err).Msg("Error getting job")
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	var result TranscriptionResult
	if lang != "" {
		for _, translation := range job.Translations {
			if translation.TargetLanguage == lang {
				err := json.Unmarshal(translation.ResultJson, &result)
				if err != nil {
					Lggr.Error().Err(err).Msg("Error decoding result")
					http.Error(w, "Error decoding result", http.StatusInternalServerError)
					return
				}
			}
		}
	} else {
		err := json.Unmarshal(job.ResultJson, &result)
		if err != nil {
			Lggr.Error().Err(err).Msg("Error decoding result")
			http.Error(w, "Error decoding result", http.StatusInternalServerError)
			return
		}
	}

	Lggr.Debug().Msgf("Result: %v", result)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v.%v", job.FileName, format))

	if format == "srt" {
		filePath := fmt.Sprintf("%v/%v-%v_%s.srt", UPLOAD_DIR, job.ID, job.FileName, result.Language)

		// Remove the file if it exists
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			os.Remove(filePath)
		}

		// Generate the new .srt file
		filePath, err := JsonToSrt(&job, &result)
		if err != nil {
			http.Error(w, "Error generating subtitles", http.StatusInternalServerError)
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "Subtitles not found", http.StatusNotFound)
			return
		}
		defer file.Close()
		w.Header().Set("Content-Type", "application/octet-stream")
		io.Copy(w, file)
	}

	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
		jsonRes, err := json.Marshal(result)
		if err != nil {
			Lggr.Error().Err(err).Msg("Error encoding result")
			http.Error(w, "Error encoding result", http.StatusInternalServerError)
			return
		}
		w.Write(jsonRes)
	}

	if format == "txt" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(result.Text))
	}
}

func DeleteJob(w http.ResponseWriter, r *http.Request) {
	var job TranscriptionJob
	id := chi.URLParam(r, "id")
	Lggr.Printf("Deleting job %v", id)
	if err := DBCon.First(&job, id).Error; err != nil {
		Lggr.Error().Err(err).Msg("Error getting job")
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Remove job files from disk
	filePath := fmt.Sprintf("%v/%v-%v", UPLOAD_DIR, job.ID, job.FileName)
	if err := os.Remove(filePath); err != nil {
		Lggr.Error().Err(err).Msg("Error deleting file")
	}

	if err := DBCon.Delete(&job).Error; err != nil {
		Lggr.Error().Err(err).Msg("Error deleting job")
		http.Error(w, "Error deleting job", http.StatusInternalServerError)
		return
	}

	// Remove all translations
	if err := DBCon.Where("transcription_job_id = ?", job.ID).Delete(&Translation{}).Error; err != nil {
		Lggr.Error().Err(err).Msg("Error deleting translations")
		http.Error(w, "Error deleting translations", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var job TranscriptionJob
	if err := DBCon.First(&job, id).Error; err != nil {
		Lggr.Error().Err(err).Msg("ERROR: Job not found")
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	filePath := fmt.Sprintf("%v/%v-%v", UPLOAD_DIR, job.ID, job.FileName)

	// Serve the file
	http.ServeFile(w, r, filePath)
}

func EditJob(w http.ResponseWriter, r *http.Request) {
	// Declare an interface to store the JSON data
	var job TranscriptionJob

	// Decode the JSON job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		Lggr.Error().Err(err).Msg("ERROR: Invalid job")
		http.Error(w, "Invalid job", http.StatusBadRequest)
		return
	}

	updateJobResult(&job)

	// Save to database
	if err := DBCon.Set("gorm:association_autoupdate", false).Save(&job).Error; err != nil {
		Lggr.Error().Err(err).Msg("ERROR: Error updating job")
		http.Error(w, "Error updating job", http.StatusInternalServerError)
		return
	}

	for _, translation := range job.Translations {
		if err := updateTranslationResult(&translation); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		Lggr.Debug().Msgf("Translation %v", translation)

		if err := DBCon.Save(&translation).Error; err != nil {
			Lggr.Error().Err(err).Msg("ERROR: Error updating translation")
			http.Error(w, "Error updating translation", http.StatusInternalServerError)
			return
		}
	}

	Lggr.Info().Msg("Job updated")
}

func updateJobResult(job *TranscriptionJob) {
	var tRes TranscriptionResult
	if err := json.Unmarshal(job.ResultJson, &tRes); err != nil {
		return
	}

	tRes.Text = ""
	for _, segment := range tRes.Segments {
		tRes.Text += segment.Text + " "
	}

	// Remove any trailing space
	tRes.Text = strings.TrimSpace(tRes.Text)

	job.ResultJson, _ = json.Marshal(tRes)
}

func updateTranslationResult(translation *Translation) error {
	var translationTRes TranscriptionResult
	if err := json.Unmarshal(translation.ResultJson, &translationTRes); err != nil {
		Lggr.Error().Err(err).Str("json_input", string(translation.ResultJson)).Msg("ERROR: Invalid translation result")
		return fmt.Errorf("invalid translation result")
	}

	for _, segment := range translationTRes.Segments {
		translationTRes.Text += segment.Text + " "
	}

	// Remove any trailing space
	translationTRes.Text = strings.TrimSpace(translationTRes.Text)
	translation.ResultJson, _ = json.Marshal(translationTRes)

	return nil
}

func TranslateTask(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	targetLanguage := chi.URLParam(r, "target_language")
	var job TranscriptionJob

	if err := DBCon.First(&job, id).Error; err != nil {
		Lggr.Error().Err(err).Msg("ERROR: TranscriptionJob not found")
		http.Error(w, "TranscriptionJob not found", http.StatusNotFound)
		return
	}

	var translation Translation
	if err := DBCon.Where("transcription_job_id = ? AND target_language = ?", id, targetLanguage).First(&translation).Error; err == nil {
		Lggr.Error().Err(err).Msg("ERROR: Translation already exists")
		http.Error(w, "Translation already exists", http.StatusConflict)
		return
	}

	newTranslation := Translation{
		TranscriptionJobID: uint(id),
		SourceLanguage:     job.Language,
		Status:             StatusPending,
		TargetLanguage:     targetLanguage,
		ResultJson:         datatypes.JSON("{}"),
	}

	if err := DBCon.Create(&newTranslation).Error; err != nil {
		Lggr.Error().Err(err).Msg("ERROR: Failed to create translation")
		http.Error(w, "Failed to create translation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTranslation)
}

func GetLanguages(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/languages", LT_ENDPOINT), nil)
	resp, err := client.Do(req)
	if err != nil {
		Lggr.Error().Err(err).Msg("Error sending request")
		return
	}
	defer resp.Body.Close()

	var result interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
