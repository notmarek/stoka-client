package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
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

type ErrorResponse struct {
	Status    string `json:"status"`
	ErrorText string `json:"error"`
}

type TokensResponse struct {
	Status       string `json:"status"`
	TokenType    string `json:"token_type"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Expiration   string `json:"expirations"`
}

func login(endpoint string, username string, password string) {
	if username == "" {
		fmt.Println("You need to provide a username to login")
		return
	}
	if password == "" {
		fmt.Println("You need to provide a password to login")
		return
	}
	res, err := http.Post(endpoint+"/na/user", "application/json", strings.NewReader(fmt.Sprintf("{\"username\":\"%s\", \"password\": \"%s\", \"identifier\":\"password\"}", username, password)))
	if err != nil {
		panic("Prolly like couldnt connect to backend or sth")
	}
	var out TokensResponse
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&out)
	fmt.Printf("%#v\n", out)
}

func main() {
	var (
		endpoint  string
		username  string
		password  string
		filepath  string
		loginflag bool
		dl        bool
	)

	flag.BoolVar(&loginflag, "login", false, "if you want to login")
	flag.BoolVar(&dl, "download", false, "if you want to dl")
	flag.StringVar(&endpoint, "endpoint", "http://127.0.0.1:1337", "Endpoint of your stoka instance")
	flag.StringVar(&filepath, "filepath", "/mnt/us/Documents/Stoka", "Where to store your documents :)")
	flag.StringVar(&username, "username", "", "your username to be used along --login")
	flag.StringVar(&password, "password", "", "your password to be used along --login")
	flag.Parse()

	if loginflag {
		login(endpoint, username, password)
	}
}
