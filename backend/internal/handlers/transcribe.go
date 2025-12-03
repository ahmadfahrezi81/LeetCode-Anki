package handlers

import (
	"leetcode-anki/backend/internal/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TranscribeHandler struct {
	llmService *services.LLMService
}

func NewTranscribeHandler() *TranscribeHandler {
	return &TranscribeHandler{
		llmService: services.NewLLMService(),
	}
}

type TranscribeResponse struct {
	Text string `json:"text"`
}

// TranscribeAudio handles POST /api/transcribe
func (h *TranscribeHandler) TranscribeAudio(c *gin.Context) {
	// Get the audio file from form
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		log.Printf("❌ Failed to get audio file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No audio file provided"})
		return
	}
	defer file.Close()

	log.Printf("🎤 Received audio file: %s (size: %d bytes)", header.Filename, header.Size)

	// Transcribe using Whisper API
	transcription, err := h.llmService.TranscribeAudio(c.Request.Context(), file, header.Filename)
	if err != nil {
		log.Printf("❌ Transcription failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transcription failed"})
		return
	}

	// Enhance the transcription with GPT-4o-mini
	enhanced, err := h.llmService.EnhanceAnswer(c.Request.Context(), transcription)
	if err != nil {
		log.Printf("⚠️ Enhancement failed, returning raw transcription: %v", err)
		// If enhancement fails, return raw transcription
		enhanced = transcription
	}

	log.Printf("✅ Transcription successful: %s", enhanced)

	// Return the enhanced text
	c.JSON(http.StatusOK, TranscribeResponse{Text: enhanced})
}
