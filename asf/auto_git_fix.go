package asf

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	fixStateFile  = "fix_state.json"
	wrongRepoFile = "wrongCommits.txt"
)

type FixState struct {
	Done map[string]bool `json:"done"`
	mu   sync.Mutex
}

func loadFixState() *FixState {
	s := &FixState{Done: make(map[string]bool)}

	file, err := os.Open(fixStateFile)
	if err != nil {
		return s
	}
	defer file.Close()

	json.NewDecoder(file).Decode(&s.Done)
	return s
}

func (s *FixState) save() {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, _ := os.Create(fixStateFile)
	defer file.Close()

	json.NewEncoder(file).Encode(s.Done)
}

func (s *FixState) markDone(repo string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Done[repo] = true
}

func RunAutoGitFix() {
	repos := readRepos(wrongRepoFile)
	total := len(repos)

	state := loadFixState()
	var doneCount int32 = int32(len(state.Done))

	jobs := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < 1; i++ {
		wg.Add(1)
		go fixWorker(i, jobs, state, total, &doneCount, &wg)
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
	fmt.Println("🎉 Fix all done")
}

func fixWorker(id int, jobs <-chan string, state *FixState, total int, doneCount *int32, wg *sync.WaitGroup) {
	defer wg.Done()

	for repo := range jobs {
		fmt.Println("======================================")
		fmt.Println("🔗 Repo:", repo)
		fmt.Println("👉 Press ENTER to continue (Ctrl+C to exit)...")

		waitEnter()

		fmt.Println("🛠️ Fixing:", repo)

		err := fixRepo(repo)
		if err != nil {
			fmt.Println("❌ Failed:", repo, err)
			continue
		}

		state.markDone(repo)
		state.save()

		current := atomic.AddInt32(doneCount, 1)

		fmt.Printf("✅ [%d/%d] (%.1f%%) Fixed: %s\n",
			current,
			total,
			float64(current)/float64(total)*100,
			repo,
		)
	}
}

func fixRepo(repoURL string) error {

	repoName := filepath.Base(repoURL)
	localPath := filepath.Join(cloneBaseDir, repoName)

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		if err := runCmd("", "git", "clone", "--depth=3", repoURL+".git", localPath); err != nil {
			return err
		}
	}

	fmt.Println("🧹 removing last commit:", repoURL)

	if !hasEnoughCommits(localPath) {
		fmt.Println("⚠️ skip, not enough commits:", repoURL)
		return nil
	}

	if err := runCmd(localPath, "git", "reset", "--hard", "HEAD~1"); err != nil {
		return err
	}

	if err := runCmd(localPath, "git", "push", "origin", "HEAD", "--force"); err != nil {
		return err
	}

	return nil
}

func waitEnter() {
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func hasEnoughCommits(localPath string) bool {
	cmd := exec.Command("git", "-C", localPath, "rev-list", "--count", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	count := strings.TrimSpace(string(out))
	return count != "0" && count != "1"
}
