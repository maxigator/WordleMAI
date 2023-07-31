package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"time"
)

var letterFrequencies = map[rune]float64{
	'e': 11.1607, 'a': 8.4966, 'r': 7.5809, 'i': 7.5448, 'o': 7.1635, 't': 6.9509, 'n': 6.6544, 's': 5.7351, 'l': 5.4893, 'c': 4.5388,
	'u': 3.6308, 'd': 3.3844, 'p': 3.1671, 'm': 3.0129, 'h': 3.0034, 'g': 2.4705, 'b': 2.0720, 'f': 1.8121, 'y': 1.7779, 'w': 1.2899,
	'k': 1.1016, 'v': 1.0074, 'x': 0.2902, 'z': 0.2722, 'j': 0.1965, 'q': 0.1962,
}

type AiWord struct {
	word  string
	usage float64
}

type Score struct {
	word  string
	score float64
}

type Guess struct {
	word   string
	scores [5]float64
}

func readLines(filename string) ([]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

func readAiWords(filename string) ([]AiWord, error) {
	lines, err := readLines(filename)
	if err != nil {
		return nil, err
	}
	aiWords := make([]AiWord, len(lines))
	for i, line := range lines {
		parts := strings.Split(line, ",")
		usage, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return nil, err
		}
		aiWords[i] = AiWord{parts[0], usage}
	}
	return aiWords, nil
}

func getColor(guess, word string, index int) (float64, error) {
	if index >= 5 || index < 0 {
		return 0.0, errors.New("Index out of bounds")
	}
	guessRunes := []rune(guess)
	wordRunes := []rune(word)
	letter := guessRunes[index]
	letterCountInWord := strings.Count(word, string(letter))
	letterCountInGuess := strings.Count(guess, string(letter))

	if wordRunes[index] == letter {
		return 4.0, nil // 4 points if the letter is in the correct position
	} else if letterCountInWord > 0 && letterCountInGuess <= letterCountInWord {
		return 1.0, nil // 2 points if the letter is in the word but in the wrong position
	} else {
		return 0.0, nil
	}
}
func checkWord(word string, guess Guess) bool {
	scores := [5]float64{0, 0, 0, 0, 0}
	for i := 0; i < 5; i++ {
		score, _ := getColor(guess.word, word, i)
		scores[i] = score
	}
	for i, val := range scores {
		if val != guess.scores[i] {
			return false
		}
	}
	return true
}

func checkAllGuesses(word string, aiguesses []Guess) bool {
	if aiguesses == nil {
		return true
	}
	for _, guess := range aiguesses {
		if !checkWord(word, guess) {
			return false
		}
	}
	return true
}

func findPossibleWords(aiguesses []Guess, aiwordArray []AiWord) []AiWord {
	var possibleWords []AiWord
	for _, aiWord := range aiwordArray {
		if checkAllGuesses(aiWord.word, aiguesses) {
			possibleWords = append(possibleWords, aiWord)
		}
	}
	sort.Slice(possibleWords, func(i, j int) bool {
		return possibleWords[i].usage > possibleWords[j].usage
	})
	return possibleWords
}

func getGuess(wordArray []string, aiwordArray []AiWord) string {
	var stat []string
	c := 0
	avg := 0
	for _, word := range wordArray {
		var guesses []Guess
		possibleWords := findPossibleWords(guesses, aiwordArray)
		if len(possibleWords) == 0 {
			// No possible words found, skip to the next word
			continue
		}
		g := Guess{
			word: possibleWords[0].word,
		}
		for i := 0; i < 5; i++ {
			score, _ := getColor(g.word, word, i)
			g.scores[i] = score
		}
		guesses = append(guesses, g)
		for len(guesses) > 0 && guesses[len(guesses)-1].word != word {
			possibleWords = findPossibleWords(guesses, aiwordArray)
			if len(possibleWords) == 0 {
				// No possible words found, break the loop
				break
			}
			g = Guess{
				word: possibleWords[0].word,
			}
			for i := 0; i < 5; i++ {
				score, _ := getColor(g.word, word, i)
				g.scores[i] = score
			}
			guesses = append(guesses, g)
		}
		var words []string
		for _, guess := range guesses {
			words = append(words, guess.word)
		}
		c += 1
		avg += len(words)
		if c%100 == 0 {
			fmt.Println(float64(avg) / float64(c))
		}
		stat = append(stat, strings.Join(words, ","))
	}
	return strings.Join(stat, "\n")
}

func calculateBestGuesses(guessFilePath string, answerFilePath string, outputFilePath string) {
	// Read and parse the guess words
	guessFileContent, _ := ioutil.ReadFile(guessFilePath)
	guessWords := strings.Split(string(guessFileContent), "\n")
	var fiveLetterGuessWords []string
	for _, word := range guessWords {
		if len(word) == 5 {
			fiveLetterGuessWords = append(fiveLetterGuessWords, word)
		}
	}

	var scores []Score
	for _, guessWord := range fiveLetterGuessWords {
		score := 0.0
		letterSet := make(map[rune]bool)
		duplicateLetters := false
		for _, letter := range guessWord {
			if _, ok := letterSet[letter]; ok {
				duplicateLetters = true
				break
			} else {
				letterSet[letter] = true
				score += letterFrequencies[letter]
			}
		}
		if !duplicateLetters {
			scores = append(scores, Score{word: guessWord, score: score})
		}
	}

	// Sort the words by score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Write the results to a new file
	var outputData []string
	for _, score := range scores {
		outputData = append(outputData, fmt.Sprintf("%s,%f", score.word, score.score))
	}
	_ = ioutil.WriteFile(outputFilePath, []byte(strings.Join(outputData, "\n")), 0644)
	fmt.Println("Word stats generated!")
}

func executeTKeyPress(wordArray []string, aiwordArray []AiWord) {
	start := time.Now()
	result := getGuess(wordArray, aiwordArray)
	err := ioutil.WriteFile("data.txt", []byte(result), 0644)
	if err != nil {
		fmt.Println("An error occurred:", err)
	} else {
		fmt.Println("File written successfully!")
	}
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Printf("time : %.0f ", elapsed.Seconds())
}

func main() {
	calculateBestGuesses("wordle-allowed-guesses.txt", "wordle-answers-alphabetical.txt", "wordle-stats.txt")
	wordArray, _ := readLines("wordle-allowed-guesses.txt")
	aiwordArray, _ := readAiWords("wordle-stats.txt")
	executeTKeyPress(wordArray, aiwordArray)
}
