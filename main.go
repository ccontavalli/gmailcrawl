package main

// This code was created from the quickstart example at this page:
//
//   https://developers.google.com/gmail/api/quickstart/go
//
// Changes made by Carlo Contavalli, ccontavalli@gmail.com.
//
// This code is released under the BSD 2-clause license, and is free
// to use and modify.
//
// Please read the LICENSE file for the full terms and conditions.

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"log"
	"net/http"
	"net/mail"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

// Client id created using https://console.developers.google.com/start/api?id=gmail
// It uniquely identifies the client to google APIs.
var clientId = "eyJpbnN0YWxsZWQiOnsiY2xpZW50X2lkIjoiNzA2MjI0NTUyMzAyLTNraGtkdnZ1dWFmYmI0YWhz" +
	"MjQ3dGg2bWo2aGhzY29mLmFwcHMuZ29vZ2xldXNlcmNvbnRlbnQuY29tIiwicHJvamVjdF9pZCI6" +
	"InBvaXNlZC1lcGlncmFtLTE2MjgwOCIsImF1dGhfdXJpIjoiaHR0cHM6Ly9hY2NvdW50cy5nb29n" +
	"bGUuY29tL28vb2F1dGgyL2F1dGgiLCJ0b2tlbl91cmkiOiJodHRwczovL2FjY291bnRzLmdvb2ds" +
	"ZS5jb20vby9vYXV0aDIvdG9rZW4iLCJhdXRoX3Byb3ZpZGVyX3g1MDlfY2VydF91cmwiOiJodHRw" +
	"czovL3d3dy5nb29nbGVhcGlzLmNvbS9vYXV0aDIvdjEvY2VydHMiLCJjbGllbnRfc2VjcmV0Ijoi" +
	"N0VoQkhqU0hfOTJ4OGt0OERlU0VxQjJZIiwicmVkaXJlY3RfdXJpcyI6WyJ1cm46aWV0Zjp3Zzpv" +
	"YXV0aDoyLjA6b29iIiwiaHR0cDovL2xvY2FsaG9zdCJdfX0="

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cfile, err := getTokenCacheFileName()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := getTokenFromFile(cfile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cfile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// getTokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func getTokenCacheFileName() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir, "gmailcrawl.json"), err
}

// Retrieves a Token from a given file path.
// Returns the retrieved Token and any read error encountered.
func getTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// Creates a file path to store the token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

var g_limit = flag.Int("limit", 0, "Limit the download to this many messages.")
var g_query = flag.String("query", "", "Query to perform to select the messages.")
var g_blacklist = flag.String("blacklist", `(mailer-daemon|password|\+bnc[A-Z0-9]*|bounce|bounce-.*|no-reply.*|do-not-reply.*|noreply.*|bounces?\+.*|prvs=.*=)@|notify@twitter.com|@((docs|.*\.bounces)\.google\.com|bounce.twitter.com)`, "Do not collect emails matching this regexp. Checked before the whitelist.")
var g_whitelist = flag.String("whitelist", "", "Only collect emails matching this regexp. Checked after the blacklist.")

func main() {
	flag.Parse()

	var err error
	var whitelist *regexp.Regexp
	if *g_whitelist != "" {
		whitelist, err = regexp.Compile(*g_whitelist)
		if err != nil {
			log.Fatalf("Could not compile -whitelist into a valid regular expression. %v", err)
		}
	}
	var blacklist *regexp.Regexp
	if *g_blacklist != "" {
		blacklist, err = regexp.Compile(*g_blacklist)
		if err != nil {
			log.Fatalf("Could not compile -blacklist into a valid regular expression. %v", err)
		}
	}

	ctx := context.Background()

	secret, err := base64.StdEncoding.DecodeString(clientId)
	if err != nil {
		log.Fatalf("Unable to decode client ID - %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/gmail-go-quickstart.json
	config, err := google.ConfigFromJSON(secret, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}

	user := "me"

	parsed := 0
	results := make(map[string]*mail.Address)
	call := srv.Users.Messages.List(user)

CallLoop:
	for {
		if *g_query != "" {
			call.Q(*g_query)
		}
		call.MaxResults(10000000)
		r, err := call.Do()
		if err != nil {
			log.Printf("Unable to retrieve labels - %v", err)
			break
		}

		if len(r.Messages) <= 0 {
			log.Print("No Messages found.")
			break
		}

		//fmt.Print("Messages:\n")
		for _, mid := range r.Messages {
			if *g_limit > 0 && parsed >= *g_limit {
				break CallLoop
			}
			parsed += 1

			//fmt.Printf("- %s, %+v\n",  mid.Id, mid)

			mgetter := gmail.NewUsersMessagesService(srv)
			call := mgetter.Get(user, mid.Id)
			call.Format("metadata")
			call.Fields("id,payload/headers")
			mymail, err := call.Do()
			if err != nil {
				continue
			}

			for _, value := range mymail.Payload.Headers {
				if value.Name != "From" && value.Name != "Cc" && value.Name != "To" && value.Name != "Bcc" &&
					value.Name != "Delivered-To" && value.Name != "Return-Path" {
					continue
				}
				addresses, _ := mail.ParseAddressList(value.Value)
				for _, address := range addresses {
					if blacklist != nil && blacklist.MatchString(address.Address) {
						continue
					}
					if whitelist != nil && !whitelist.MatchString(address.Address) {
						continue
					}
					key := strings.ToLower(address.Address)
					found := results[key]
					if found != nil {
						if len(found.Name) <= 0 && len(address.Name) > 0 {
							found.Name = address.Name
						}
					} else {
						results[address.Address] = address
					}
				}
			}
		}
		if len(r.NextPageToken) <= 0 {
			break
		}
		call = srv.Users.Messages.List(user)
		call.PageToken(r.NextPageToken)
	}

	for _, value := range results {
		fmt.Printf("%s %s\n", value.Address, value.Name)
	}
}
