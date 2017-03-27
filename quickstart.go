package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
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

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gmail-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/gmail-go-quickstart.json
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}

	user := "me"

	results := make(map[string]*mail.Address)
	call := srv.Users.Messages.List(user)
	for {
		call.Fields("messages(id,payload/headers),nextPageToken")
		call.MaxResults(10000000)
		r, err := call.Do()
		if err != nil {
			log.Fatalf("Unable to retrieve labels. %v", err)
			break
		}

		// Fields: From, To, Cc, Bcc
		if len(r.Messages) > 0 {
			//fmt.Print("Messages:\n")
			for _, mid := range r.Messages {
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
					if value.Name == "From" || value.Name == "Cc" || value.Name == "To" || value.Name == "Bcc" || value.Name == "Delivered-To" || value.Name == "Return-Path" {
						addresses, _ := mail.ParseAddressList(value.Value)
						for _, address := range addresses {
							found := results[address.Address]
							if found != nil {
								if len(found.Name) <= 0 && len(address.Name) > 0 {
									found.Name = address.Name
								}
							} else {
								results[address.Address] = address
							}
						}
					}

					//fmt.Printf("HDR: %+v\n", value)
				}
				//fmt.Printf("%+v\n", results)
				//  for _, value := range results {
				//    fmt.Printf("%s %s\n", value.Address, value.Name)
				//  }
				//        return
			}
		} else {
			fmt.Print("No Messages found.")
			break
		}
		if len(r.NextPageToken) <= 0 {
			break
		}
		newcall := srv.Users.Messages.List(user)
		newcall.PageToken(r.NextPageToken)
		call = newcall
	}

	for _, value := range results {
		fmt.Printf("%s %s\n", value.Address, value.Name)
	}
}
