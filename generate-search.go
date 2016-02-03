package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/mrvdot/golang-utils"
	"github.com/toqueteos/webbrowser"
)

// Search words in files
type Search struct {
	Path  string
	Words []string
}

// Page Title
type Page struct {
	Title string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

const root string = "."

const filenameCsv string = ".data.csv"
const filenamejson string = ".data.json"

func stripchars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return r
		}
		return -1
	}, str)
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

func replace(path string) {
	copy := []string{}
	r := `(<script(\s|\S)*?<\/script>)|(<style(\s|\S)*?<\/style>)|(<!--(\s|\S)*?-->)|(<\/?(\s|\S)*?>)|(nbsp;)|((?:\s)\s)|(png)|(jpeg)|(jpg)|(mpg)|(\\u0026)|(\n)|(\v)|(\r)|(\0)|(\t)|(n°)
		|(à)|(wbe)|(_)`
	regex, err := regexp.Compile(r)
	if err != nil {
		return // there was a problem with the regular expression.
	}
	c, _ := readLines(path)
	for _, v := range c {
		reg := regex.ReplaceAllString(v, " ")
		slug := utils.GenerateSlug(reg)
		regex1, _ := regexp.Compile(`((\-){1,})|(\b\w{1}\b)`)
		reg = regex1.ReplaceAllString(slug, " ")
		t := stripchars(reg, `?,.!/©*@#~()$+"'&}]|:;[{²`)
		s := strings.TrimSpace(t)
		// fmt.Println(s)

		normalize := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
		normStr1, _, _ := transform.String(normalize, s)
		// fmt.Println(normStr1)

		if len(v) > 0 {
			copy = append(copy, normStr1)
		}
	}

	// fmt.Println(cleaned, "\n")

	j := strings.Replace(strings.Join((copy), " "), " ", ",", -1)
	// fmt.Println(j)
	regex2, err := regexp.Compile(`((\,){2,})`)
	j1 := regex2.ReplaceAllString(j, ",")
	// fmt.Println(j1)
	j2 := strings.Split(j1, ",")

	cleaned := []string{}

	for _, value := range j2 {
		if !stringInSlice(value, cleaned) {
			cleaned = append(cleaned, value)
		}
	}
	createCsv(path, filenameCsv, strings.Join(cleaned, ","))
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}

func uniq(list []string) []string {
	uniqueSet := make(map[string]bool, len(list))
	for _, x := range list {
		uniqueSet[x] = true
	}
	result := make([]string, 0, len(uniqueSet))
	for x := range uniqueSet {
		result = append(result, x)
	}
	return result
}

func removeDuplicates(elements []string) []string {
	result := []string{}

	for i := 0; i < len(elements); i++ {
		// Scan slice for a previous element of the same value.
		exists := false
		for v := 0; v < i; v++ {
			if elements[v] == elements[i] {
				exists = true
				break
			}
		}
		// If no previous element exists, append this one.
		if !exists {
			result = append(result, elements[i])
		}
	}
	return result
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func createCsv(path, file string, tab string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	s := fmt.Sprintf("%s,%s\n", path, tab)
	n, err := io.WriteString(f, s)
	if err != nil {
		fmt.Println(n, err)
	}
	f.Close()
}

func walkpath(path string, f os.FileInfo, err error) error {
	if !f.IsDir() && filepath.Ext(f.Name()) == ".html" {
		replace(path)
	}
	return nil
}

func deleteFile(f string) {
	// delete file
	var err = os.Remove(f)
	check(err)
}

func createFile(f string) {
	// detect if file exists
	var _, err = os.Stat(f)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(f)
		check(err)
		defer file.Close()
	} else {
		var err = os.Remove(f)
		check(err)
		createFile(f)
	}
}

func jsonWrite(path string) error {
	csvFile, _ := os.Open(path)

	reader := csv.NewReader(csvFile)

	reader.FieldsPerRecord = -1

	csvData, err := reader.ReadAll()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var oneRecord Search
	var allRecords []Search

	for _, each := range csvData {
		oneRecord.Path = each[0]
		oneRecord.Words = each[1:]
		allRecords = append(allRecords, oneRecord)
	}

	jsonData, err := json.Marshal(allRecords)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	jsonFile, err := os.Create(filenamejson)

	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	jsonFile.Write(jsonData)
	fmt.Println("Fichier .data.json créé le ", time.Now())

	csvFile.Close()
	deleteFile(filenameCsv)

	return err
}

func main() {
	go func() {
		fmt.Println("Création du .json le ", time.Now())
		fmt.Println("Veuillez patienter quelques instants. Une fois fini, un message vous l'indiquera.")
		createFile(filenameCsv)

		if err := filepath.Walk(root, walkpath); err == nil {
			if err := jsonWrite(filenameCsv); err == nil {
				webbrowser.Open("http://localhost:8080")
			}
		}
	}()
	log.Println("Le serveur est bien démarré")
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))

	http.ListenAndServe(":8080", nil)

}
