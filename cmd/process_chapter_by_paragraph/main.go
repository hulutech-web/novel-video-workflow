package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"novel-video-workflow/pkg/tools/indextts2"

	"go.uber.org/zap"
)

func main() {
	// Define paths
	inputChapterPath := "./input/幽灵客栈/chapter_08/chapter_08.txt"
	referenceAudioPath := "./assets/ref_audio/ref.m4a"
	outputDir := "./output/幽灵客栈/chapter_08"

	// Read chapter content
	fmt.Println("Reading chapter content...")
	content, err := ioutil.ReadFile(inputChapterPath)
	if err != nil {
		panic(fmt.Sprintf("Cannot read chapter file: %v", err))
	}

	chapterText := string(content)
	fmt.Printf("Successfully read chapter content, length: %d characters\n", len(chapterText))

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(fmt.Sprintf("Cannot create output directory: %v", err))
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Cannot initialize logger: %v", err))
	}
	defer logger.Sync()

	// Use Indextts2 client
	client := indextts2.NewIndexTTS2Client(logger, "http://localhost:7860")

	// Split text by paragraph (using \n\n\n as separator)
	paragraphs := strings.Split(chapterText, "\n\n\n")
	fmt.Printf("Detected %d paragraphs\n", len(paragraphs))

	// Process each paragraph
	for i, paragraph := range paragraphs {
		// Skip empty paragraphs
		if strings.TrimSpace(paragraph) == "" {
			fmt.Printf("Skipping empty paragraph %d\n", i+1)
			continue
		}

		// Generate audio filename
		outputAudioPath := fmt.Sprintf("%s/paragraph_%02d.wav", outputDir, i+1)
		fmt.Printf("Processing paragraph %d, length: %d characters, content: %.50s...\n", i+1, len(paragraph), paragraph)

		err = client.GenerateTTSWithAudio(referenceAudioPath, paragraph, outputAudioPath)
		if err != nil {
			fmt.Printf("Warning: Failed to process paragraph %d: %v\n", i+1, err)
			continue
		}

		fmt.Printf("Generated audio: %s\n", outputAudioPath)

		// Add brief delay to avoid too frequent requests
		time.Sleep(2 * time.Second)
	}

	fmt.Println("All paragraphs processed!")
}

// splitLongParagraph splits long paragraph by sentences
func splitLongParagraph(text string, maxLength int) []string {
	var result []string

	// Split by sentence-ending punctuation
	sentences := splitBySentences(text)

	currentChunk := ""
	for _, sentence := range sentences {
		// Check if adding current sentence would exceed max length
		if len(currentChunk)+len(sentence) <= maxLength {
			if currentChunk != "" {
				currentChunk += "\n" + sentence
			} else {
				currentChunk = sentence
			}
		} else {
			// If current chunk is not empty, save it first
			if currentChunk != "" {
				result = append(result, currentChunk)
			}

			// If single sentence exceeds max length, force split
			if len(sentence) > maxLength {
				parts := splitLongSentence(sentence, maxLength)
				for _, part := range parts {
					result = append(result, part)
				}
				currentChunk = ""
			} else {
				currentChunk = sentence
			}
		}
	}

	// Add final chunk
	if currentChunk != "" {
		result = append(result, currentChunk)
	}

	return result
}

// splitBySentences splits text by sentence-ending punctuation
func splitBySentences(text string) []string {
	var sentences []string

	// Define sentence-ending punctuation
	endMarkers := []string{"。", "！", "？", ".", "!", "?"}

	currentPos := 0
	for i := 0; i < len(text); i++ {
		char := string(text[i])

		// Check if it's a sentence-ending punctuation
		isEndMarker := false
		for _, marker := range endMarkers {
			if char == marker {
				isEndMarker = true
				break
			}
		}

		if isEndMarker {
			// Find consecutive end punctuation
			endPos := i + 1
			for endPos < len(text) && isEndMarkerChar(string(text[endPos])) {
				endPos++
			}

			sentence := strings.TrimSpace(text[currentPos:endPos])
			if sentence != "" {
				sentences = append(sentences, sentence)
			}

			currentPos = endPos
			i = endPos - 1 // -1 because loop will increment i
		}
	}

	// Add remaining part
	remaining := strings.TrimSpace(text[currentPos:])
	if remaining != "" {
		sentences = append(sentences, remaining)
	}

	return sentences
}

// isEndMarkerChar checks if character is sentence-ending punctuation
func isEndMarkerChar(char string) bool {
	endMarkers := []string{"。", "！", "？", ".", "!", "?"}
	for _, marker := range endMarkers {
		if char == marker {
			return true
		}
	}
	return false
}

// splitLongSentence splits overly long sentence
func splitLongSentence(sentence string, maxLength int) []string {
	var result []string

	// Simple split by length, try to split at word boundaries
	for len(sentence) > maxLength {
		// Find suitable position to split (prefer at space or comma)
		splitPos := maxLength

		// Try to split at comma
		for i := maxLength; i > maxLength-50 && i > 0; i-- {
			if string(sentence[i]) == "，" || string(sentence[i]) == "," {
				splitPos = i + 1
				break
			}
		}

		result = append(result, sentence[:splitPos])
		sentence = sentence[splitPos:]
	}

	if sentence != "" {
		result = append(result, sentence)
	}

	return result
}
