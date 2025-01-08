package main

import (
	"flag"
	"fmt"
	pkgerrs "github.com/pkg/errors"
	"os"
	"strings"
	"sync"
)

type wc struct {
	lines       bool
	words       bool
	characters  bool
	fileNames   []string
	resultSlice []result
}

func main() {

	//flag.String("l", "", "word count")
	//log.Logger{} //TODO: lookup a generic logger which you can use

	wc, err := fetchCommandLineArguments()
	if err != nil {
		fmt.Println(err)
	}

	if err := wc.performCommandOperations(); err != nil {
		fmt.Println(err)
		os.Exit(1) //non-zero exit code
	}
	os.Exit(0)
}

func fetchCommandLineArguments() (wc, error) {
	lines := flag.Bool("l", false, "get number of lines")
	words := flag.Bool("w", false, "get number of words")
	characters := flag.Bool("c", false, "get number of characters")
	flag.Parse()
	args := flag.Args()

	var inputFiles []string
	for _, arg := range args {
		inputFiles = append(inputFiles, arg)
	}
	if len(inputFiles) == 0 {
		//TODO: handle input from stdin
	}

	return wc{lines: *lines, words: *words, characters: *characters, fileNames: inputFiles}, nil
}

type result struct {
	numOfLines int
	numOfWords int
	numOfChars int
	error      error
	fileName   string
}

func (w wc) performCommandOperations() error {

	totalLines, totalWords, totalCharacters := 0, 0, 0

	//create a chan of type result with length w.fileNames
	resultSlice := make([]result, len(w.fileNames))
	w.resultSlice = resultSlice

	waitGrp := sync.WaitGroup{}

	for idx, fileName := range w.fileNames {
		idx := idx
		fileName := fileName
		waitGrp.Add(1)
		w.findLineWordAndCharacters(idx, fileName, &waitGrp)
	}
	waitGrp.Wait()
	for _, result := range w.resultSlice {
		w.logOutput(result)
		if result.error == nil {
			totalLines += result.numOfLines
			totalWords += result.numOfWords
			totalCharacters += result.numOfChars
		}
	}
	fmt.Fprintln(os.Stdout, fmt.Sprintf("%8d %8d %8d %s", totalLines, totalWords, totalCharacters, "total"))
	return nil
}

func (w wc) findLineWordAndCharacters(idx int, fileName string, waitGrp *sync.WaitGroup) {
	defer waitGrp.Done()
	open, err := os.Open(fileName)
	if err != nil {
		pathError := err.(*os.PathError)
		//TODO: find a better way for error handling here
		w.resultSlice[idx] = result{
			error: pkgerrs.New("./wc:" + " " + fileName + ": " + pathError.Op + ": " + pathError.Err.Error()),
		}
		return
	}

	fileInfo, err := open.Stat()
	if err != nil || fileInfo.IsDir() {
		w.resultSlice[idx] = result{
			error: pkgerrs.New("./wc:" + " " + fileName + ": " + "read: Is a directory"),
		}
		return
	}
	defer open.Close()
	numOfLines, numOfWords, numOfCharacters := 0, 0, 0
	buffer := make([]byte, 1024)
	for {
		readBytes, _ := open.Read(buffer)
		if readBytes == 0 {
			break
		}
		numOfLines += w.calculateLineCount(buffer, readBytes)
		numOfWords += w.calculateWordCount(buffer, readBytes)
		numOfCharacters += w.calculateCharacterCount(buffer, readBytes)

	}
	w.resultSlice[idx] = w.createResult(numOfLines, numOfWords, numOfCharacters, fileName)
	return
}

func (w wc) calculateCharacterCount(buffer []byte, bytesToRead int) int {
	return bytesToRead
}

func (w wc) calculateWordCount(buffer []byte, bytesToRead int) int {
	//TODO: fix the word count bug
	return len(strings.Fields(string(buffer[:bytesToRead])))

}

func (w wc) calculateLineCount(buffer []byte, bytesToRead int) int {
	return strings.Count(string(buffer[:bytesToRead]), "\n")
}

func (w wc) createResult(lines int, words int, characters int, fileName string) result {
	var output string
	var resultStr result
	resultStr.fileName = fileName
	if w.lines {
		//output += fmt.Sprintf("%8d", lines)
		resultStr.numOfLines = lines
	}
	if w.words {
		//output += fmt.Sprintf("%8d", words)
		resultStr.numOfWords = words
	}
	if w.characters {
		//output += fmt.Sprintf("%8d", characters)
		resultStr.numOfChars = characters
	}
	if !w.lines && !w.words && !w.characters {
		output += fmt.Sprintf("%8d %8d %8d", lines, words, characters)
	}
	output += " " + fileName
	return resultStr
	//return output
}

func (w wc) logOutput(resultStruct result) {
	if resultStruct.error != nil {
		fmt.Fprintln(os.Stderr, resultStruct.error.Error())
		return
	}
	var output string
	if w.lines {
		output += fmt.Sprintf("%8d", resultStruct.numOfLines)
	}
	if w.words {
		output += fmt.Sprintf("%8d", resultStruct.numOfWords)
	}
	if w.characters {
		output += fmt.Sprintf("%8d", resultStruct.numOfChars)
	}
	if !w.lines && !w.words && !w.characters {
		output += fmt.Sprintf("%8d %8d %8d", resultStruct.numOfLines,
			resultStruct.numOfWords,
			resultStruct.numOfChars)
	}
	output += " " + resultStruct.fileName
	fmt.Fprintln(os.Stdout, output)
}
