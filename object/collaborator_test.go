package object

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/casbin/casbin-oa/proxy"
	"github.com/casbin/casbin-oa/util"
	"github.com/google/go-github/v74/github"
)

func listDirectCollaborators(org string) (map[string][]string, error) {
	client := util.GetClient()
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	result := make(map[string][]string)

	i := 1
	for {
		repos, resp, err := client.Repositories.ListByOrg(context.Background(), org, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for org %s: %v", org, err)
		}

		for _, repo := range repos {
			repoName := repo.GetName()
			if repoName == "" {
				continue
			}

			collabs, _, err := client.Repositories.ListCollaborators(
				context.Background(),
				org,
				repoName,
				&github.ListCollaboratorsOptions{Affiliation: "direct"},
			)
			if err != nil {
				fmt.Printf("[%s] failed to list collaborators for repo %s: %v\n", time.Now().Format("2006-01-02 15:04:05"), repoName, err)
				continue
			}

			var usernames []string
			for _, c := range collabs {
				usernames = append(usernames, c.GetLogin())
			}

			repoURL := fmt.Sprintf("https://github.com/%s/%s/settings/access", org, repoName)
			result[repoURL] = usernames

			if len(usernames) > 0 {
				fmt.Printf("[%s] [%d] Repo: %s, Collaborators: %v\n", time.Now().Format("2006-01-02 15:04:05"), i, repoURL, usernames)
			}
			i += 1
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return result, nil
}

func TestListDirectCollaborators(t *testing.T) {
	InitConfig()
	proxy.InitHttpClient()

	org := "casbin"
	_, err := listDirectCollaborators(org)
	if err != nil {
		t.Fatalf("Error listing collaborators: %v\n", err)
	}
}
