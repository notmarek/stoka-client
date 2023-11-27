package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
)

type AcquiredBooks struct {
	Hashes []string `json:"filehashes"`
}

type ConfigFile struct {
	Tokens   TokensResponse `json:"tokens"`
	Endpoint string         `json:"endpoint"`
	FilePath string         `json:"filepath"`
}

type FileType struct {
	Name string `json:"name"`
}
type Book struct {
	Id       int      `json:"id"`
	Title    string   `json:"title"`
	Hash     string   `json:"hash"`
	FileType FileType `json:"file_type"`
}

type Response[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

type ErrorResponse struct {
	Status    string `json:"status"`
	ErrorText string `json:"error"`
}

type TokensResponse struct {
	Status       string `json:"status"`
	TokenType    string `json:"token_type"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Expiration   string `json:"expiration"`
}

func loadJson[T any](path string) (obj T, err error) {
	in, err := os.Open(path)
	if err != nil {
		return *new(T), err
	}
	defer in.Close()
	var out T
	data, _ := io.ReadAll(in)
	_ = json.Unmarshal(data, &out)
	return out, nil
}

func saveJson(obj any, path string) (err error) {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	j, _ := json.MarshalIndent(obj, "", "\t")
	out.WriteString(string(j))
	return nil
}

func downloadFile(filepath string, url string, token string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
func download_book(endpoint string, token string, file_path string, id int) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/book/%d", endpoint, id), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ := client.Do(req)
	defer res.Body.Close()
	var out Response[Book]
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&out)
	downloadFile(fmt.Sprintf("%s/%s.%s", file_path, out.Data.Title, out.Data.FileType.Name), fmt.Sprintf("%s/api/book/%d/dl", endpoint, id), token)

}
func download_books(endpoint string, token string, seen_hashes []string, file_path string) {
	books := get_books(endpoint, token)
	for _, book := range books {
		if !slices.Contains(seen_hashes, book.Hash) {
			seen_hashes = append(seen_hashes, book.Hash)
			download_book(endpoint, token, file_path, book.Id)
		} else {
			fmt.Printf("Skipping \"%#v\"", book)
		}
	}
	saveJson(AcquiredBooks{Hashes: seen_hashes}, file_path+"/.seen.json")
}

func get_books(endpoint string, token string) []Book {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", endpoint+"/api/books", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ := client.Do(req)
	defer res.Body.Close()

	var out Response[[]Book]
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&out)

	// save the data to a config file
	// fmt.Printf("%#v\n", out)
	return out.Data

}

func login(endpoint string, username string, password string) TokensResponse {
	if username == "" {
		fmt.Println("You need to provide a username to login")
		// return nil
	}
	if password == "" {
		fmt.Println("You need to provide a password to login")
		// return nil
	}
	res, err := http.Post(endpoint+"/na/user", "application/json", strings.NewReader(fmt.Sprintf("{\"username\":\"%s\", \"password\": \"%s\", \"identifier\":\"password\"}", username, password)))
	if err != nil {
		panic("Prolly like couldnt connect to backend or sth")
	}
	if res.StatusCode == 200 {
		var out TokensResponse
		defer res.Body.Close()
		decoder := json.NewDecoder(res.Body)
		decoder.Decode(&out)

		// save the data to a config file
		// fmt.Printf("%#v\n", out)
		return out
	} else {
		var out ErrorResponse
		defer res.Body.Close()
		decoder := json.NewDecoder(res.Body)
		decoder.Decode(&out)
		fmt.Printf("Error: %s\n", out.ErrorText)
		panic("nope")
	}
}

func main() {
	var (
		tokens      TokensResponse
		endpoint    string
		username    string
		password    string
		filepath    string
		confpath    string
		seen_hashes []string
		loginflag   bool
		dl          bool
	)

	flag.BoolVar(&loginflag, "login", false, "if you want to login")
	flag.BoolVar(&dl, "download", false, "if you want to dl")
	flag.StringVar(&endpoint, "endpoint", "https://stoka.notmarek.com", "Endpoint of your stoka instance")
	flag.StringVar(&confpath, "confpath", "/mnt/us/stoka.json", "Where to store the config file.")
	flag.StringVar(&filepath, "filepath", "/mnt/us/Stoka", "Where to store your documents :)")
	flag.StringVar(&username, "username", "", "your username to be used along --login")
	flag.StringVar(&password, "password", "", "your password to be used along --login")
	flag.Parse()
	os.MkdirAll(filepath, os.ModePerm)

	if loginflag {
		tokens = login(endpoint, username, password)
		config := ConfigFile{
			Tokens:   tokens,
			FilePath: filepath,
			Endpoint: endpoint,
		}
		saveJson(config, confpath)
		fmt.Println("Config saved!")
		return
	} else {
		conf, err := loadJson[ConfigFile](confpath)
		if err != nil {
			fmt.Println("Couldn't load config file!")
		} else {
			endpoint = conf.Endpoint
			tokens = conf.Tokens
			filepath = conf.FilePath
		}

		sh, err := loadJson[AcquiredBooks](filepath + "/.seen.json")
		if err != nil {
			fmt.Println("Couldn't load seen hashes!")
			seen_hashes = []string{}
		} else {
			seen_hashes = sh.Hashes
		}
	}

	if dl {
		download_books(endpoint, tokens.Token, seen_hashes, filepath)
	}
}
