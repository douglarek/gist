package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"github.com/howeyc/gopass"
)

var (
	client   *github.Client
	gistFile = filepath.Join(os.Getenv("HOME"), ".gist")
	ctx      = context.Background()
)

func init() {
	dt, err := ioutil.ReadFile(gistFile)
	if err != nil {
		log.Printf("*WARNING*: `%v`, you are Anonymous!", err)
		client = github.NewClient(nil)
	} else {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(dt)})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}
}

// Gist stands for gist related ops.
type Gist struct {
	*github.Client
}

// Create makes a gist.
func (g *Gist) Create(description string, anonymous, public bool, files ...string) (err error) {
	fs := make(map[github.GistFilename]github.GistFile, len(files))
	for _, v := range files {
		dat, err := ioutil.ReadFile(v)
		if err != nil {
			return err
		}
		c := string(dat)
		vv := strings.Split(v, "/")
		name := vv[len(vv)-1]
		fs[github.GistFilename(name)] = github.GistFile{Filename: &name, Content: &c}
	}
	g0 := &github.Gist{Files: fs, Public: &public, Description: &description}
	if anonymous {
		*g.Client = *github.NewClient(nil)
	}
	g0, _, err = g.Gists.Create(ctx, g0)
	if err == nil {
		fmt.Println(*g0.HTMLURL)
	}
	return
}

// List gets user's gists.
func (g *Gist) List(user string, public bool) (err error) {
	opt := &github.GistListOptions{
		ListOptions: github.ListOptions{
			PerPage: 20,
		},
	}
	for {
		gs, resp, err := g.Gists.List(ctx, user, opt)
		if err != nil {
			return err
		}
		for _, i := range gs {
			if public && *i.Public {
				continue
			}

			for fn := range i.Files {
				fmt.Printf("%-64s%s\n", *i.HTMLURL, fn)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}
	return
}

// Get querys a single gist detail.
func (g *Gist) Get(id string) (err error) {
	g0, _, err := g.Gists.Get(ctx, id)
	if err != nil {
		return
	}
	fmt.Println(strings.Repeat("-", 100))
	for _, f := range g0.Files {
		fmt.Printf("%v\t%v\n\n%v\n", *f.Filename, *f.Size, *f.Content)
		fmt.Println(strings.Repeat("-", 100))
	}
	return
}

// Delete deletes gaven gists by ids.
func (g *Gist) Delete(id ...string) error {
	c := make(chan error, len(id))
	for _, i := range id {
		go func(id string) {
			_, err := g.Gists.Delete(ctx, id)
			if err == nil {
				fmt.Printf("<id: %s> has been deleted ...\n", id)
			}
			c <- err
		}(i)
	}
	for i := 0; i < len(id); i++ {
		if err := <-c; err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

// Token is a GitHub token entry.
type Token struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
	App struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		ClientID string `json:"client_id"`
	} `json:"app"`
	Token          string      `json:"token"`
	HashedToken    string      `json:"hashed_token"`
	TokenLastEight string      `json:"token_last_eight"`
	Note           string      `json:"note"`
	NoteURL        interface{} `json:"note_url"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Scopes         []string    `json:"scopes"`
	Fingerprint    interface{} `json:"fingerprint"`
}

func ask() (user, pass string) {
	fmt.Print("GitHub username: ")
	if _, err := fmt.Scan(&user); err != nil {
		return
	}
	fmt.Print("GitHub password: ")
	p, err := gopass.GetPasswdMasked()
	if err != nil {
		return
	}
	pass = string(p)
	return

}

func basicRequest(user, pass, otp string) (*http.Request, error) {
	fp := time.Now().Nanosecond()
	note := fmt.Sprintf(`{"note": "gist","scopes":["gist"],"fingerprint":"%v"}`, fp)
	url := "https://api.github.com/authorizations"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(note)))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")
	if len(otp) != 0 {
		req.Header.Set("X-GitHub-OTP", otp)
	}
	return req, nil
}

func token(user, pass string) (err error) {
	req, err := basicRequest(user, pass, "")
	if err != nil {
		return nil
	}
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if strings.HasPrefix(resp.Header.Get("X-Github-Otp"), "required") {
		var code string
		fmt.Print("GitHub OTP: ")
		fmt.Scan(&code)
		req, err := basicRequest(user, pass, code)
		if err != nil {
			return nil
		}
		resp, err = client.Do(req)
	}
	if err != nil {
		return
	}

	if sc := resp.StatusCode; sc == http.StatusUnauthorized {
		return errors.New(http.StatusText(sc))
	}

	defer resp.Body.Close()

	var t Token
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return err
	}

	if err := ioutil.WriteFile(gistFile, []byte(t.Token), 0644); err != nil {
		return err
	}
	fmt.Println("success ...")
	return nil
}
