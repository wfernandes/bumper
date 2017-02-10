package gitter

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/wfernandes/bumper/colors"
)

type Git struct {
	config  *config
	commits []*Commit
}

type Commit struct {
	Hash    string
	Subject string
	// TODO: use a pointer to story here vs denormalizing
	StoryID   int
	StoryName string
	Accepted  bool
}

func New(opts ...GitOption) *Git {

	config := setupConfig(opts)

	return &Git{
		config: config,
	}
}

func (g *Git) Commits() []*Commit {
	return g.commits
}

func (g *Git) Start() error {

	fmt.Printf("Bumping the following range of commits: %s\n\n",
		colors.ExtraRed(g.config.commitRange))

	g.hashes()

	if len(g.commits) == 0 {
		return errors.New("no commits")
	}
	g.subjects()
	g.storyIDs()
	return nil
}

func (g *Git) hashes() {
	commitRange := g.config.commitRange
	cmd := exec.Command("git", "log", "--pretty=format:%H", commitRange)
	out := &bytes.Buffer{}
	cmd.Stdout = out
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Unable to run command: %s", err)
	}

	commits := make([]*Commit, 0)
	br := bufio.NewReader(out)
	for {
		bytes, _, err := br.ReadLine()
		if err != nil {
			break
		}
		commit := &Commit{Hash: string(bytes)}
		commits = append(commits, commit)
	}
	g.commits = commits
}

func (g *Git) subjects() {
	for _, c := range g.commits {
		cmd := exec.Command("git", "show", "--no-patch", "--pretty=format:%s", c.Hash)
		out := &bytes.Buffer{}
		cmd.Stdout = out
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Unable to run command: %s", err)
		}
		c.Subject = out.String()
	}
}

func (g *Git) storyIDs() {
	for _, c := range g.commits {
		cmd := exec.Command("git", "show", "--no-patch", "--pretty=format:%B", c.Hash)
		out := &bytes.Buffer{}
		cmd.Stdout = out
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Unable to run command: %s", err)
		}
		c.StoryID = getStoryID(out.String())
	}
}

func getStoryID(body string) int {
	reStoryID := regexp.MustCompile(`\[#(\d+)\]`)
	result := reStoryID.FindStringSubmatch(body)
	if len(result) < 2 {
		return 0
	}
	storyID := result[1]
	id, err := strconv.Atoi(storyID)
	if err != nil {
		return 0
	}
	return id
}

type GitOption func(c *config)

func setupConfig(opts []GitOption) *config {

	config := &config{
		commitRange: "master..release-elect",
	}

	for _, o := range opts {
		o(config)
	}

	return config

}

type config struct {
	commitRange string
}

func WithCommitRange(cr string) func(c *config) {
	return func(c *config) {
		c.commitRange = cr
	}
}
