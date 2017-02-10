package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wfernandes/bumper/colors"
	"github.com/wfernandes/bumper/gitter"
)

var (
	commitRange string
	trackerKey  string
	projectID   int
)

type Story struct {
	ID    int    `json:"id"`
	State string `json:"current_state"`
	Name  string `json:"name"`
}

var princeQuotes = []string{
	"A strong spirit transcends rules.",
	"You can always tell when the groove is working or not.",
	"Everyone has a rock bottom.",
	"The internet's completely over.",
	"So tonight we gonna party like it's 1999.",
}

func randomPrinceQuote() string {
	return princeQuotes[rand.Int()%len(princeQuotes)]
}

func main() {
	// validate tracker api key and project id were set
	trackerKey = os.Getenv("TRACKER_KEY")
	if trackerKey == "" {
		log.Fatalf("Invalid Tracker Key")
	}
	var err error
	projectID, err = strconv.Atoi(os.Getenv("PROJECT_ID"))
	if err != nil {
		log.Fatalf("Invalid Project ID: %s", err)
	}

	// figure out ranges default to master..release-elect
	flag.Parse()

	git := gitter.New(
		gitter.WithCommitRange(commitRange),
	)

	// collect all commits to check
	err = git.Start()
	if err != nil {
		fmt.Println("There are no commits to bump!")
		return
	}

	commits := git.Commits()

	// in parallel check those commits against the tracker api
	isAccepted(commits)

	// output commit informations
	var maxSubject int
	for _, c := range commits {
		if len(c.Subject) > maxSubject {
			maxSubject = len(c.Subject)
		}
	}
	for _, c := range commits {
		mark := colors.Red("✗")
		if c.Accepted {
			mark = colors.Green("✓")
		}
		if c.StoryID == 0 {
			mark = colors.Prince("✓")
		}
		subject := c.Subject
		if len(subject) > 50 {
			subject = subject[:47] + "..."
		}
		subject = padRight(subject, " ", maxSubject)
		storyID := strconv.Itoa(c.StoryID)
		if c.StoryID == 0 {
			storyID = colors.Prince(getDancer())
		}
		storyName := c.StoryName
		if c.StoryName == "" {
			storyName = colors.Prince(randomPrinceQuote())
		}
		fmt.Println(
			mark,
			colors.Yellow(c.Hash[:8]),
			colors.Grey(subject),
			colors.Blue(storyID),
			colors.Grey(storyName),
		)
	}
	fmt.Println()

	// reverse the commits before we find the bump commit
	reversed := make([]*gitter.Commit, len(commits))
	for i, c := range commits {
		reversed[len(commits)-1-i] = c
	}
	commits = reversed

	// find bump commit and output it
	bumpHash := findBump(commits)
	if bumpHash == "" {
		fmt.Println("There are no commits to bump!")
		return
	}

	fmt.Println("This is the commit you should bump to: ")
	fmt.Println(colors.ExtraRed(bumpHash))
}

func isAccepted(commits []*gitter.Commit) {
	const urlTemplate = "https://www.pivotaltracker.com/services/v5/projects/%d/stories?%s"

	// build filter
	// TODO: dedupe these so we don't hit the api with multiples of the same
	// story ids
	ids := make([]string, 0)
	for _, c := range commits {
		ids = append(ids, strconv.Itoa(c.StoryID))
	}
	filter := strings.Join(ids, " OR ")
	v := url.Values{}
	v.Set("filter", filter)

	url := fmt.Sprintf(urlTemplate, projectID, v.Encode())

	// make request
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// unmarshal bytes
	stories := make([]Story, len(commits))
	err = json.Unmarshal(bytes, &stories)
	if err != nil {
		log.Fatal(err)
	}

	// update commits
	for _, c := range commits {
		if c.StoryID == 0 {
			c.Accepted = true
			continue
		}
		for _, s := range stories {
			if c.StoryID == s.ID {
				c.StoryName = s.Name
				if s.State == "accepted" {
					c.Accepted = true
				}
			}
		}
	}
}

func findBump(commits []*gitter.Commit) string {
	invalid := make(map[int]bool)
	firstUnaccepted := -1
	bumpHash := ""

	// find invalid index
	for i, c := range commits {
		if !c.Accepted {
			firstUnaccepted = i
			break
		}
	}

	// return early if all stories are accepted
	if firstUnaccepted == -1 {
		// this shouldn't panic since len(commits) is always > 0
		return commits[len(commits)-1].Hash
	}

	// record invalid stories
	for _, c := range commits[firstUnaccepted:] {
		if c.Accepted && c.StoryID != 0 {
			invalid[c.StoryID] = true
		}
	}

	// find last commit that is accpeted and not invalid
	for _, c := range commits[:firstUnaccepted] {
		_, ok := invalid[c.StoryID]
		if ok {
			break
		}
		bumpHash = c.Hash
	}
	return bumpHash
}

func padRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

var dancerToggle bool

func getDancer() string {
	if dancerToggle {
		dancerToggle = false
		return "┏ (･o･)┛♪"
	}
	dancerToggle = true
	return "♪┗ (･o･)┓"
}

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.StringVar(
		&commitRange,
		"commit-range",
		"master..release-elect",
		"Specifies the commit range to consider bumping.",
	)
}
