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
)

const (
	fixStateFile  = "fix_state.json"
	wrongRepoFile = "revert_wrong_commits.txt"
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
	doneCount := len(state.Done)

	for _, repo := range repos {
		if state.Done[repo] {
			fmt.Println("⏭️ skip:", repo)
			continue
		}

		fmt.Println("======================================")
		fmt.Println("🔗 Repo:", repo)
		fmt.Println("Ctrl+C to exit...")

		waitEnter()

		fmt.Println("🛠️ Fixing:", repo)

		err := fixRepo(repo)
		if err != nil {
			fmt.Println("❌ Failed:", repo, err)
			continue
		}

		state.markDone(repo)
		state.save()

		doneCount++

		fmt.Printf("✅ [%d/%d] (%.1f%%) Fixed: %s\n",
			doneCount,
			total,
			float64(doneCount)/float64(total)*100,
			repo,
		)
	}

	fmt.Println("🎉 Fix all done")
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
