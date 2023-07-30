package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
)

type AiWord struct {
	word  string
	usage int
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
	letter := guess[index]
	letterCountInWord := strings.Count(word, string(letter))
	letterCountInGuess := strings.Count(guess, string(letter))
	if word[index] == letter {
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
		fmt.Println(strings.Join(words, ","))
		stat = append(stat, strings.Join(words, ","))
	}
	return strings.Join(stat, ",")
}

func executeTKeyPress(wordArray []string, aiwordArray []AiWord) {
	result := getGuess(wordArray, aiwordArray)
	err := ioutil.WriteFile("data.json", []byte(result), 0644)
	if err != nil {
		fmt.Println("An error occurred:", err)
	} else {
		fmt.Println("File written successfully!")
	}
}

func main() {
	wordArray, _ := readLines("wordle-allowed-guesses.txt")
	aiwordArray, _ := readAiWords("wordle-stats.txt")
	executeTKeyPress(wordArray, aiwordArray)
}
