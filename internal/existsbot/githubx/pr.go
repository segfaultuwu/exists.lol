package githubx

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-github/v64/github"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/registry"
	"golang.org/x/oauth2"
)

type Client struct {
	owner string
	repo  string
	gh    *github.Client
}

type CreateDomainPROptions struct {
	DiscordUsername string
	DiscordID       string
	GitHubUsername  string
	Subdomain       string
	RecordType      string
	Value           string

	ExtraRecords map[string]string
}

type GitHubFileRequest struct {
	Message string `json:"message"`
	Content string `json:"content"`
	SHA     string `json:"sha,omitempty"`
}

func (c *Client) UserExists(ctx context.Context, username string) (bool, error) {
	_, resp, err := c.gh.Users.Get(ctx, username)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func New(token, owner, repo string) *Client {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})

	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		owner: owner,
		repo:  repo,
		gh:    github.NewClient(tc),
	}
}

func (c *Client) CreateDomainPR(ctx context.Context, opts CreateDomainPROptions) (string, error) {
	baseBranch := "main"
	path := fmt.Sprintf("domains/%s.json", opts.Subdomain)

	_, _, resp, err := c.gh.Repositories.GetContents(
		ctx,
		c.owner,
		c.repo,
		path,
		&github.RepositoryContentGetOptions{
			Ref: baseBranch,
		},
	)
	if err == nil {
		return "", fmt.Errorf("subdomain %q already exists", opts.Subdomain)
	}

	if resp == nil || resp.StatusCode != 404 {
		return "", fmt.Errorf("check domain file: %w", err)
	}

	newBranch := fmt.Sprintf("bot/add-%s-%d", opts.Subdomain, time.Now().Unix())

	ref, _, err := c.gh.Git.GetRef(ctx, c.owner, c.repo, "refs/heads/"+baseBranch)
	if err != nil {
		return "", fmt.Errorf("get main ref: %w", err)
	}

	_, _, err = c.gh.Git.CreateRef(ctx, c.owner, c.repo, &github.Reference{
		Ref: github.String("refs/heads/" + newBranch),
		Object: &github.GitObject{
			SHA: ref.Object.SHA,
		},
	})
	if err != nil {
		return "", fmt.Errorf("create branch: %w", err)
	}

	file := registry.NewDomainFile(
		opts.DiscordUsername,
		opts.DiscordID,
		opts.GitHubUsername,
		opts.RecordType,
		opts.Value,
	)

	raw, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal domain file: %w", err)
	}

	content := append(raw, '\n')

	_, _, err = c.gh.Repositories.CreateFile(ctx, c.owner, c.repo, path, &github.RepositoryContentFileOptions{
		Message: github.String("add " + opts.Subdomain + ".exists.lol"),
		Content: content,
		Branch:  github.String(newBranch),
	})
	if err != nil {
		return "", fmt.Errorf("create domain file: %w", err)
	}

	body := fmt.Sprintf(
		"Requested from Discord by `%s` (`%s`).\n\nGitHub: `@%s`\nSubdomain: `%s.exists.lol`\nRecord: `%s`\nValue: `%s`",
		opts.DiscordUsername,
		opts.DiscordID,
		opts.GitHubUsername,
		opts.Subdomain,
		opts.RecordType,
		opts.Value,
	)

	pr, _, err := c.gh.PullRequests.Create(ctx, c.owner, c.repo, &github.NewPullRequest{
		Title: github.String("Add " + opts.Subdomain + ".exists.lol"),
		Head:  github.String(newBranch),
		Base:  github.String(baseBranch),
		Body:  github.String(body),
	})
	if err != nil {
		return "", fmt.Errorf("create pull request: %w", err)
	}

	return pr.GetHTMLURL(), nil
}
