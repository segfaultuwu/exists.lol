package github

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-github/v64/github"
	"github.com/segfaultuwu/exists.lol/internal/registry"
	"golang.org/x/oauth2"
)

// Client wraps GitHub API client with repository context
type Client struct {
	owner string
	repo  string
	gh    *github.Client
}

// DomainRequest represents a request to create a new domain via PR
type DomainRequest struct {
	Subdomain     string              `json:"subdomain"`
	Owner         registry.Owner      `json:"owner"`
	Records       map[string][]string `json:"records"`
	PRTitle       string              `json:"pr_title,omitempty"`
	PRDescription string              `json:"pr_description,omitempty"`
}

// PRResult contains information about a created pull request
type PRResult struct {
	URL       string    `json:"url"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Branch    string    `json:"branch"`
	CreatedAt time.Time `json:"created_at"`
}

// New creates a new GitHub client
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

// UserExists checks if a GitHub user exists
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

// CreateDomainPR creates a new domain via pull request
func (c *Client) CreateDomainPR(ctx context.Context, req DomainRequest) (*PRResult, error) {
	baseBranch := "main"
	filePath := fmt.Sprintf("domains/%s.json", req.Subdomain)

	// Check if domain file already exists
	_, _, resp, err := c.gh.Repositories.GetContents(
		ctx,
		c.owner,
		c.repo,
		filePath,
		&github.RepositoryContentGetOptions{
			Ref: baseBranch,
		},
	)
	if err == nil {
		return nil, fmt.Errorf("subdomain %q already exists", req.Subdomain)
	}

	if resp == nil || resp.StatusCode != 404 {
		return nil, fmt.Errorf("check domain file: %w", err)
	}

	// Create a new branch
	newBranch := fmt.Sprintf("bot/add-%s-%d", req.Subdomain, time.Now().Unix())

	ref, _, err := c.gh.Git.GetRef(ctx, c.owner, c.repo, "refs/heads/"+baseBranch)
	if err != nil {
		return nil, fmt.Errorf("get main ref: %w", err)
	}

	_, _, err = c.gh.Git.CreateRef(ctx, c.owner, c.repo, &github.Reference{
		Ref: github.String("refs/heads/" + newBranch),
		Object: &github.GitObject{
			SHA: ref.Object.SHA,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create branch: %w", err)
	}

	// Create domain file content
	domain := registry.DomainFile{
		Owner:   req.Owner,
		Records: req.Records,
	}

	data, err := json.MarshalIndent(domain, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal domain file: %w", err)
	}

	content := string(data) + "\n"

	// Create the file in the new branch
	_, _, err = c.gh.Repositories.CreateFile(ctx, c.owner, c.repo, filePath, &github.RepositoryContentFileOptions{
		Message: github.String("add " + req.Subdomain + ".exists.lol"),
		Content: []byte(content),
		Branch:  github.String(newBranch),
	})
	if err != nil {
		return nil, fmt.Errorf("create domain file: %w", err)
	}

	// Create pull request
	prTitle := req.PRTitle
	if prTitle == "" {
		prTitle = "Add " + req.Subdomain + ".exists.lol"
	}

	prBody := req.PRDescription
	if prBody == "" {
		prBody = fmt.Sprintf(
			"Requested from Discord by `%s` (`%s`).\n\nGitHub: `@%s`\nSubdomain: `%s.exists.lol`",
			req.Owner.Username,
			req.Owner.DiscordID,
			req.Owner.GitHubUsername,
			req.Subdomain,
		)
	}

	pr, _, err := c.gh.PullRequests.Create(ctx, c.owner, c.repo, &github.NewPullRequest{
		Title: github.String(prTitle),
		Head:  github.String(newBranch),
		Base:  github.String(baseBranch),
		Body:  github.String(prBody),
	})
	if err != nil {
		return nil, fmt.Errorf("create pull request: %w", err)
	}

	return &PRResult{
		URL:       pr.GetHTMLURL(),
		Number:    pr.GetNumber(),
		Title:     pr.GetTitle(),
		Branch:    newBranch,
		CreatedAt: pr.GetCreatedAt().Time,
	}, nil
}

// GetFile retrieves a file from the repository
func (c *Client) GetFile(ctx context.Context, path, ref string) ([]byte, error) {
	content, _, _, err := c.gh.Repositories.GetContents(
		ctx,
		c.owner,
		c.repo,
		path,
		&github.RepositoryContentGetOptions{
			Ref: ref,
		},
	)
	if err != nil {
		return nil, err
	}

	if content == nil {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	data, err := content.GetContent()
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

// ListFiles lists all domain files in the repository
func (c *Client) ListFiles(ctx context.Context, path string) ([]*github.RepositoryContent, error) {
	_, contents, _, err := c.gh.Repositories.GetContents(
		ctx,
		c.owner,
		c.repo,
		path,
		&github.RepositoryContentGetOptions{},
	)
	if err != nil {
		return nil, err
	}

	return contents, nil
}
