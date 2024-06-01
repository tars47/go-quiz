package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jwalton/gchalk"
)

type record []string

type result struct {
	correct   int
	showWrong string
	done      bool
}

var csvFileName string
var duration int

func init() {

	flag.StringVar(&csvFileName, "csv", "problems.csv", "a csv file name in the format like problems.csv")
	flag.IntVar(&duration, "timer", 30, "specifies the time limit of the quiz in seconds")
}

func main() {

	flag.Parse()

	if err := validateCsv(csvFileName); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	records, err := readCsvFile(csvFileName)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Press Enter to start %d sec timed quiz!", duration)

	var qres result

	rchan := make(chan result)

	fmt.Scanln()

	timer := time.NewTimer(time.Duration(duration) * time.Second)

	go quiz(records, rchan)

outer:
	for {
		select {
		case <-timer.C:
			fmt.Print(gchalk.Red("\n\nTimes Up!!"))
			break outer
		case qres = <-rchan:
			if qres.done {
				break outer
			}
		}
	}

	fmt.Println("------------------------------------------------------------")
	fmt.Println(
		gchalk.Blue(
			gchalk.Underline(
				fmt.Sprintf("You scored %d out of %d \n", qres.correct, len(records)))))
	fmt.Print(qres.showWrong)
	fmt.Println("------------------------------------------------------------")

}

func validateCsv(fileName string) error {
	strs := strings.Split(fileName, ".")
	if len(strs) != 2 || strs[1] != "csv" {
		return errors.New("Invalid file format, should have format problems.csv")
	}
	return nil
}

func readCsvFile(fileName string) ([]record, error) {
	file, err := os.Open(fileName) // For read access.
	if err != nil {
		return nil, errors.New("Could not open " + fileName + ". Please check if this is a vilid file")
	}

	r := csv.NewReader(file)
	records := make([]record, 0, 12)

	for {
		r, err := r.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		var question string
		var answer string

		for i, v := range r {
			if len(r) != i+1 {
				question += v
			} else {
				answer = v
			}

		}

		records = append(records, record{question, answer})

	}
	return records, nil
}

func quiz(records []record, rchan chan<- result) {
	var correct int
	var showWrong string = ``
	input := bufio.NewScanner(os.Stdin)

	for i, v := range records {
		question, answer := v[0], v[1]

		fmtq := fmt.Sprintf("Problem#%d \t: %v =", i+1, question)
		fmt.Print(fmtq)

		rchan <- result{correct: correct, showWrong: showWrong + fmt.Sprintln(fmtq, "\t", gchalk.Green(answer))}

		input.Scan()

		uAns := strings.TrimSpace(input.Text())

		if uAns == answer {
			correct++
		} else {
			showWrong += fmt.Sprintln(fmtq, gchalk.Red(uAns), "\t", gchalk.Green(answer))
		}
		rchan <- result{correct: correct, showWrong: showWrong}
	}

	rchan <- result{correct: correct, showWrong: showWrong, done: true}

}
