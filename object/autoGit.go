package object

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type Repo struct {
	Name string `json:"name"`
}

const (
	org          = "apache"
	keyword      = "casbin"
	repoListFile = "repos.txt"
	configFile   = "asf.yaml"
	cloneBaseDir = "./cloned_repos"
	stateFile    = "state.json"
	commitMsg    = "chore: add config file"
	githubPrefix = "https://github.com/"
	workerCount  = 3
)

type State struct {
	Done map[string]bool `json:"done"`
	mu   sync.Mutex
}

func loadState() *State {
	s := &State{Done: make(map[string]bool)}

	file, err := os.Open(stateFile)
	if err != nil {
		return s
	}
	defer file.Close()

	json.NewDecoder(file).Decode(&s.Done)
	return s
}

func (s *State) save() {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, _ := os.Create(stateFile)
	defer file.Close()

	json.NewEncoder(file).Encode(s.Done)
}

func (s *State) markDone(repo string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Done[repo] = true
}

func fetchRepos() {
	fmt.Println("Fetching repository list...")
	var allNames []string
	page := 1

	for {
		url := fmt.Sprintf("https://api.github.com/orgs/%s/repos?per_page=100&page=%d", org, page)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("request error:", err)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var repos []Repo
		if err := json.Unmarshal(body, &repos); err != nil {
			fmt.Println("json error:", err)
			return
		}

		if len(repos) == 0 {
			break
		}

		for _, r := range repos {
			if strings.Contains(strings.ToLower(r.Name), strings.ToLower(keyword)) {
				allNames = append(allNames, r.Name)
			}
		}

		page++
	}

	sort.Slice(allNames, func(i, j int) bool {
		return strings.ToLower(allNames[i]) < strings.ToLower(allNames[j])
	})

	f, err := os.Create(repoListFile)
	if err != nil {
		fmt.Println("file error:", err)
		return
	}
	defer f.Close()

	for _, name := range allNames {
		f.WriteString(fmt.Sprintf("%s/%s\n", org, name))
	}

	fmt.Printf("Done! %d repos written to %s\n", len(allNames), repoListFile)
	fmt.Println("Repository list fetched successfully.")
}

func RunAutoGitAddFile() {

	fetchRepos()

	os.MkdirAll(cloneBaseDir, os.ModePerm)

	repos := readRepos(repoListFile)
	total := len(repos)

	state := loadState()

	var doneCount int32 = int32(len(state.Done))

	jobs := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(i, jobs, state, total, &doneCount, &wg)
	}

	for _, repo := range repos {
		if state.Done[repo] {
			fmt.Println("⏭️ skip:", repo)
			continue
		}
		jobs <- repo
	}
	close(jobs)

	wg.Wait()
	fmt.Println("🎉 All done")
}

func worker(id int, jobs <-chan string, state *State, total int, doneCount *int32, wg *sync.WaitGroup) {
	defer wg.Done()

	for repo := range jobs {
		fmt.Println("🚀 Processing:", repo)

		err := processRepo(repo)
		if err != nil {
			fmt.Println("❌ Failed:", repo, err)
			continue
		}

		state.markDone(repo)
		state.save()

		current := atomic.AddInt32(doneCount, 1)

		fmt.Printf("✅ [%d/%d] (%.1f%%) Done: %s\n",
			current,
			total,
			float64(current)/float64(total)*100,
			repo,
		)
	}
}

func processRepo(repo string) error {
	repoURL := githubPrefix + repo + ".git"
	repoName := filepath.Base(repo)
	localPath := filepath.Join(cloneBaseDir, repoName)

	cleanup := false

	defer func() {
		if cleanup {
			if err := os.RemoveAll(localPath); err != nil {
				fmt.Println("⚠️ cleanup failed:", localPath, err)
			} else {
				fmt.Println("🧹 cleaned:", localPath)
			}
		}
	}()

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		if err := runCmd("", "git", "clone", "--depth=1", repoURL, localPath); err != nil {
			return err
		}
	}

	destFile := filepath.Join(localPath, filepath.Base(configFile))
	if err := copyFile(configFile, destFile); err != nil {
		return err
	}

	if err := runCmd(localPath, "git", "add", "."); err != nil {
		return err
	}

	runCmd(localPath, "git", "commit", "-m", commitMsg)

	if err := runCmd(localPath, "git", "push"); err != nil {
		return err
	}

	cleanup = true
	return nil
}

func readRepos(file string) []string {
	var repos []string

	f, _ := os.Open(file)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		r := strings.TrimSpace(scanner.Text())
		if r != "" {
			repos = append(repos, r)
		}
	}
	return repos
}

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
