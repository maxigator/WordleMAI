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

type AiWord struct {
	word  string
	usage int
}

type Score struct {
	word  string
	score int
}

type Guess struct {
	word   string
	scores [5]int
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
		usage, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		aiWords[i] = AiWord{parts[0], usage}
	}
	return aiWords, nil
}

func getColor(guess, word string, index int) (int, error) {
	if index >= 5 || index < 0 {
		return 0, errors.New("Index out of bounds")
	}
	guessRunes := []rune(guess)
	wordRunes := []rune(word)
	letter := guessRunes[index]
	letterCountInWord := strings.Count(word, string(letter))
	letterCountInGuess := strings.Count(guess, string(letter))

	if wordRunes[index] == letter {
		return 2, nil
	} else if letterCountInWord > 0 && letterCountInGuess <= letterCountInWord {
		return 1, nil
	} else {
		return 0, nil
	}
}

func checkWord(word string, guess Guess) bool {
	scores := [5]int{0, 0, 0, 0, 0}
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
		g := Guess{
			word: possibleWords[0].word,
		}
		for i := 0; i < 5; i++ {
			score, _ := getColor(g.word, word, i)
			g.scores[i] = score
		}
		guesses = append(guesses, g)
		for guesses[len(guesses)-1].word != word {
			possibleWords = findPossibleWords(guesses, aiwordArray)
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

	// Read and parse the answer words
	answerFileContent, _ := ioutil.ReadFile(answerFilePath)
	answerWords := strings.Split(string(answerFileContent), "\n")
	var fiveLetterAnswerWords []string
	for _, word := range answerWords {
		if len(word) == 5 {
			fiveLetterAnswerWords = append(fiveLetterAnswerWords, word)
		}
	}

	var scores []Score
	for _, guessWord := range fiveLetterGuessWords {
		score := 0
		for _, answerWord := range fiveLetterAnswerWords {
			for i := 0; i < 5; i++ {
				tempScore, _ := getColor(guessWord, answerWord, i)
				score += tempScore
			}
		}
		scores = append(scores, Score{word: guessWord, score: score})
	}

	// Sort the words by score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Write the results to a new file
	var outputData []string
	for _, score := range scores {
		outputData = append(outputData, fmt.Sprintf("%s,%d", score.word, score.score))
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
