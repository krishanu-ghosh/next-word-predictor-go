package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

type Predictor struct {
	Chain map[string]map[string]int
}

func NewPredictor() *Predictor {
	return &Predictor{Chain: make(map[string]map[string]int)}
}

func clean(word string) string {
	return strings.ToLower(strings.Trim(word, ".,!?;:\"()"))
}

// Inside TrainFromFile, only keep words that appear more than once
// OR use a prune function after training
func (p *Predictor) Prune() {
	for prefix, nexts := range p.Chain {
		for word, freq := range nexts {
			// Hardcoded threshold: remove anything seen only once
			if freq < 2 {
				delete(nexts, word)
			}
		}
		// Clean up the prefix if it no longer has any associated next words
		if len(p.Chain[prefix]) == 0 {
			delete(p.Chain, prefix)
		}
	}
}

func (p *Predictor) TrainFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	wordCount := 0
	startTime := time.Now()

	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line)

		if len(words) < 3 {
			continue
		}

		var w1, w2 string
		for _, rawWord := range words {
			w3 := clean(rawWord)
			if w3 == "" {
				continue
			}

			if w1 != "" && w2 != "" {
				prefix := w1 + " " + w2
				if _, ok := p.Chain[prefix]; !ok {
					p.Chain[prefix] = make(map[string]int)
				}
				p.Chain[prefix][w3]++
			}
			w1, w2 = w2, w3

			// Progress Logging
			wordCount++
			if wordCount%1000000 == 0 {
				fmt.Printf("Processed %d million words... (Time: %v)\n", wordCount/1000000, time.Since(startTime).Round(time.Second))
			}
		}
	}
	fmt.Printf("Training complete. Total words: %d\n", wordCount)
	return scanner.Err()
}

// finds the single most likely next word
func (p *Predictor) PredictNext(w1, w2 string) string {
	prefix := clean(w1) + " " + clean(w2)
	nextWords, exists := p.Chain[prefix]
	if !exists {
		return ""
	}
	bestWord, maxFreq := "", -1
	for word, freq := range nextWords {
		if freq > maxFreq {
			maxFreq = freq
			bestWord = word
		}
	}
	return bestWord
}

// creates a sequence of N words starting from a 2-word seed
func (p *Predictor) Generate(w1, w2 string, length int) string {
	currentW1, currentW2 := clean(w1), clean(w2)
	result := []string{currentW1, currentW2}

	for i := 0; i < length; i++ {
		next := p.PredictNext(currentW1, currentW2)
		if next == "" {
			break
		}
		result = append(result, next)
		currentW1, currentW2 = currentW2, next
	}
	return strings.Join(result, " ")
}

func (p *Predictor) Save(filename string) error {
	file, _ := os.Create(filename)
	defer file.Close()
	return gob.NewEncoder(file).Encode(p.Chain)
}

func (p *Predictor) Load(filename string) error {
	file, _ := os.Open(filename)
	defer file.Close()
	return gob.NewDecoder(file).Decode(&p.Chain)
}

func main() {
	debug.SetGCPercent(100)
	rand.NewSource(time.Now().UnixNano())
	model := NewPredictor()
	modelFile := "model.gob"
	// corpusFile := "news.2007.en.shuffled"

	//Load existing model if it exists
	if _, err := os.Stat(modelFile); err == nil {
		fmt.Println("Loading existing model knowledge...")
		if err := model.Load(modelFile); err != nil {
			fmt.Printf("Warning: Failed to load existing model: %v\n", err)
		}
	}

	//Check for corpus and train/append
	// if _, err := os.Stat(corpusFile); err == nil {
	// 	fmt.Printf("Training/Appending knowledge from %s...\n", corpusFile)
	// 	if err := model.TrainFromFile(corpusFile); err != nil {
	// 		fmt.Printf("Critical: Training failed: %v\n", err)
	// 		return
	// 	}

	// 	fmt.Println("Optimizing memory (Pruning)...")
	// 	model.Prune()

	// 	fmt.Println("Saving updated model...")
	// 	if err := model.Save(modelFile); err != nil {
	// 		fmt.Printf("Critical: Failed to save model: %v\n", err)
	// 	}
	// } else {
	// 	fmt.Println("No corpus file found. Using existing knowledge only.")
	// }

	// fmt.Printf("Current Model Stats: %d unique prefixes indexed.\n", len(model.Chain))

	//Dynamic Input Loop
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n--- Multiple Next Words Predictor ---")
	fmt.Println("Enter 2 words to start (or 'exit' to quit):")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "exit" {
			break
		}

		words := strings.Fields(input)
		if len(words) < 2 {
			fmt.Println("Please enter at least 2 words.")
			continue
		}

		w1, w2 := words[len(words)-2], words[len(words)-1]
		prediction := model.Generate(w1, w2, 10)
		fmt.Printf("Generated Sequence: ... %s\n", prediction)
	}
}
