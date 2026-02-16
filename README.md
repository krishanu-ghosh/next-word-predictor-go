# Go Markov Text Predictor

A lightweight text prediction engine based on a 2-order Markov Chain. It ingests large text corpora, builds a probability map of word sequences, and generates text based on user input.

## Features

* **2-Word Context:** Uses a sliding window of two words (trigrams) to predict the third.
* **Persistence:** Serializes the trained chain to disk (`model.gob`) using Go's `encoding/gob` for fast reloading.
* **Memory Optimization:** Includes a `Prune()` method to remove low-frequency transitions (frequency < 2) to reduce model size.
* **Interactive Shell:** CLI-based interface for testing predictions in real-time.

## Prerequisites

* Go 1.25+ (I'm using 1.25.5)
* A text file for training (corpus)

## Usage

### 1. Training
The code provided has the training logic commented out by default. To train the model:

1.  Place your text corpus (e.g., `news.txt`) in the root directory.
2.  Open `main.go` and **uncomment lines 153-169**.
3.  Update the `corpusFile` variable (line 143) to match your filename.
4.  Run the program:
    ```bash
    go run main.go
    ```
    *This will create `model.gob`.*

### 2. Running Predictions
Once `model.gob` exists, run the application:

```bash
go run main.go
```

# Technical Architecture

This project implements a **2nd-Order Markov Chain** text generator in Go. It utilizes a probabilistic approach (a "trigram" model) to predict the next word based on the context of the previous two words.

Below is a detailed breakdown of the internal components.

## 1. The Core Data Structure
The "brain" of the program is stored in the `Predictor` struct:

```go
type Predictor struct {
    Chain map[string]map[string]int
}
```
Outer Map (map[string]...): The key is the Prefix. This is a string composed of two words joined by a space (e.g., "united states").

Inner Map (...map[string]int): The key is the Suffix (the potential next word). 

The value (int) is the Frequency (how many times this suffix appeared after that specific prefix).

Visual Representation: If the text "the cat sat on the cat mat" is processed, the memory structure looks like this: Prefix (Key)Inner Map (Value)

Explanation: "the cat"{"sat": 1, "mat": 1}After "the cat", both "sat" and "mat" appeared once."cat sat"{"on": 1}"on" is the only word seen after "cat sat"."sat on"{"the": 1}"the" is the only word seen after "sat on".

## 2. Data Cleaning (clean)
Before processing, every word passes through a clean() helper function:

Normalization: Converts strings to lowercase (strings.ToLower). This ensures "The" and "the" are treated as the same token.Sanitization: Trims punctuation (.,!?;:"()) from the edges. For example, "Apple," becomes "apple". 

## 3. The Training Phase (TrainFromFile)
This function ingests a text file to build the Chain.

Buffered Reading: Uses bufio.Scanner with a manually set 1MB buffer. This prevents crashes on lines longer than the default 64KB buffer (common in minified text files).

The Sliding Window: The logic relies on three variables: w1 (Word N-2), w2 (Word N-1), and w3 (Current Word).

Step A: Read w3.

Step B: If w1 and w2 are not empty, combine them to form the prefix key: prefix := w1 + " " + w2.

Step C: Access Chain[prefix]. If it doesn't exist, initialize the inner map.

Step D: Increment the count for w3: p.Chain[prefix][w3]++.

Step E: Shift the window. w1 becomes w2, and w2 becomes w3.
## 4. Memory Optimization (Prune)
Markov chains grow exponentially with large corpora. The Prune function reduces RAM usage by removing "noise".

Thresholding: Iterates through every suffix in the chain. If a word has a frequency less than 2 (meaning it was only seen once), it is deleted.

Dead Key Removal: If a prefix loses all its potential next words (the inner map becomes empty), the prefix itself is deleted from the outer map. 
## 5. Prediction Logic (PredictNext)
This function determines the single "best" next word.

Input: Takes two words (w1, w2).

Lookup: Generates the key (clean(w1) + " " + clean(w2)).

Selection Algorithm (Greedy/Deterministic):Iterates through all candidates in the inner map.Tracks the maxFreq.Returns the word with the highest frequency.

Note: This implementation is deterministic. If "the cat" is followed by "sat" (50 times) and "ate" (49 times), it will always pick "sat". It does not use weighted random selection.
## 6. Sequence Generation (Generate)
Chains predictions together to form sentences.

Seed: Starts with a user-provided 2-word seed.Loop: Runs for length iterations (set to 10 in main).

Feed-Forward: Calls PredictNext(currentW1, currentW2).If no prediction is found (empty string), the loop breaks early. Appends the result to the output slice.

Updates the seed: currentW1 becomes currentW2, currentW2 becomes next.
## 7. Persistence (Save / Load)
The program saves its "brain" so it doesn't have to re-read the text file every time.

Format: Uses encoding/gob (Go Binary).Why Gob? It is significantly faster and more compact than JSON for serializing native Go maps. It preserves the exact structure of the Chain map on disk.
## 8. Main Execution Flow
GC Tuning: debug.SetGCPercent(100) tells the Garbage Collector to run only when heap size doubles.

Model Loading: Checks for model.gob. If found, loads binary data directly into memory.

Training: (Currently commented out) Logic exists to read news.2007.en.shuffled if enabled.

Interactive Loop: Reads user input from Stdin.Expects at least 2 words. Extracts the last two words to use as the seed.Prints the generated 10-word continuation.