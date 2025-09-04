package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type GitHubFile struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
}

type FrontMatter struct {
	Title        string     `json:"title"`
	Date         time.Time  `json:"date"`
	LastModified *time.Time `json:"last_modified,omitempty"`
	Slug         string     `json:"slug"`
}

type BlogConfig struct {
	IsLocal    bool
	LocalPath  string
	GitHost    string // github, gitlab, etc.
	GitRepo    string // owner/repo
	GitHubAPI  string
	RawBaseURL string
}

func getBlogConfig() BlogConfig {
	blogSource := os.Getenv("BLOG_SOURCE")
	if blogSource == "" {
		blogSource = "github:ethanthoma/blogs"
	}

	if strings.HasPrefix(blogSource, "/") || strings.HasPrefix(blogSource, "~/") || strings.HasPrefix(blogSource, "./") || strings.HasPrefix(blogSource, "../") {
		localPath := blogSource

		if strings.HasPrefix(localPath, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Printf("Blog: Failed to get home directory: %v", err)
			} else {
				localPath = filepath.Join(homeDir, localPath[2:])
			}
		}

		return BlogConfig{
			IsLocal:   true,
			LocalPath: localPath,
		}
	}

	if !strings.Contains(blogSource, ":") {
		log.Printf("Blog: Invalid git source format '%s'. Expected format: 'host:owner/repo' (e.g., 'github:owner/repo')", blogSource)
		return BlogConfig{}
	}

	parts := strings.SplitN(blogSource, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		log.Printf("Blog: Invalid git source format '%s'. Expected format: 'host:owner/repo'", blogSource)
		return BlogConfig{}
	}

	gitHost := parts[0]
	gitRepo := parts[1]

	var apiURL, rawURL string
	switch gitHost {
	case "github":
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/contents", gitRepo)
		rawURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/main/", gitRepo)
	case "gitlab":
		apiURL = fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/repository/tree", strings.ReplaceAll(gitRepo, "/", "%2F"))
		rawURL = fmt.Sprintf("https://gitlab.com/%s/-/raw/main/", gitRepo)
	default:
		log.Printf("Blog: Unsupported git host '%s', defaulting to GitHub", gitHost)
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/contents", gitRepo)
		rawURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/main/", gitRepo)
		gitHost = "github"
	}

	return BlogConfig{
		IsLocal:    false,
		GitHost:    gitHost,
		GitRepo:    gitRepo,
		GitHubAPI:  apiURL,
		RawBaseURL: rawURL,
	}
}

func GetPostsFromGitHub() ([]Post, error) {
	config := getBlogConfig()

	if !config.IsLocal && config.GitHost == "" {
		return nil, fmt.Errorf("invalid blog source configuration")
	}

	if config.IsLocal {
		return getPostsFromLocal(config.LocalPath)
	}
	return getPostsFromGitAPI(config)
}

func getPostsFromLocal(localPath string) ([]Post, error) {
	log.Printf("Blog: fetching posts from local directory %s...", localPath)

	files, err := os.ReadDir(localPath)
	if err != nil {
		log.Printf("Blog: failed to read local directory: %v", err)
		return nil, err
	}

	var posts []Post
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") && file.Name() != "README.md" {
			post, err := fetchPostFromLocal(localPath, file.Name())
			if err != nil {
				log.Printf("Blog: failed to fetch local post %s: %v", file.Name(), err)
				continue
			}
			posts = append(posts, post)
		}
	}

	for i := 0; i < len(posts)-1; i++ {
		for j := i + 1; j < len(posts); j++ {
			if posts[i].Date.Before(posts[j].Date) {
				posts[i], posts[j] = posts[j], posts[i]
			}
		}
	}

	log.Printf("Blog: fetched %d posts from local directory", len(posts))
	return posts, nil
}

func getPostsFromGitAPI(config BlogConfig) ([]Post, error) {
	log.Printf("Blog: fetching posts from %s...", config.GitHost)

	resp, err := http.Get(config.GitHubAPI)
	if err != nil {
		log.Printf("Blog: failed to fetch file list from %s: %v", config.GitHost, err)
		return nil, err
	}
	defer resp.Body.Close()

	var files []GitHubFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		log.Printf("Blog: failed to decode %s file list: %v", config.GitHost, err)
		return nil, err
	}

	var posts []Post
	for _, file := range files {
		if strings.HasSuffix(file.Name, ".md") && file.Name != "README.md" {
			post, err := fetchPostFromGitAPI(config, file.Name)
			if err != nil {
				log.Printf("Blog: failed to fetch %s post %s: %v", config.GitHost, file.Name, err)
				continue
			}
			posts = append(posts, post)
		}
	}

	for i := 0; i < len(posts)-1; i++ {
		for j := i + 1; j < len(posts); j++ {
			if posts[i].Date.Before(posts[j].Date) {
				posts[i], posts[j] = posts[j], posts[i]
			}
		}
	}

	log.Printf("Blog: fetched %d posts from %s", len(posts), config.GitHost)
	return posts, nil
}

func GetPostFromGitHub(slug string) (Post, error) {
	config := getBlogConfig()
	filename := slug + ".md"

	if !config.IsLocal && config.GitHost == "" {
		return Post{}, fmt.Errorf("invalid blog source configuration")
	}

	if config.IsLocal {
		return fetchPostFromLocal(config.LocalPath, filename)
	}
	return fetchPostFromGitAPI(config, filename)
}

func fetchPostFromLocal(localPath, filename string) (Post, error) {
	log.Printf("Blog: fetching local post %s...", filename)

	filePath := filepath.Join(localPath, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return Post{}, fmt.Errorf("failed to read local file %s: %v", filename, err)
	}

	post, err := parseMarkdownWithFrontmatter(string(content), localPath, filename)
	if err != nil {
		return Post{}, fmt.Errorf("failed to parse markdown: %v", err)
	}

	if post.Slug == "" {
		post.Slug = strings.TrimSuffix(filename, ".md")
	}

	post.TLDR = extract_tldr(post.Content)

	log.Printf("Blog: fetched local post %s", post.Slug)
	return post, nil
}

func fetchPostFromGitAPI(config BlogConfig, filename string) (Post, error) {
	log.Printf("Blog: fetching %s post %s...", config.GitHost, filename)

	url := config.RawBaseURL + filename
	resp, err := http.Get(url)
	if err != nil {
		return Post{}, fmt.Errorf("failed to fetch file %s: %v", filename, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Post{}, fmt.Errorf("file %s not found (status: %d)", filename, resp.StatusCode)
	}

	var content bytes.Buffer
	_, err = content.ReadFrom(resp.Body)
	if err != nil {
		return Post{}, fmt.Errorf("failed to read file content: %v", err)
	}

	post, err := parseMarkdownWithFrontmatter(content.String(), "", filename)
	if err != nil {
		return Post{}, fmt.Errorf("failed to parse markdown: %v", err)
	}

	if post.Slug == "" {
		post.Slug = strings.TrimSuffix(filename, ".md")
	}

	post.TLDR = extract_tldr(post.Content)

	log.Printf("Blog: fetched %s post %s", config.GitHost, post.Slug)
	return post, nil
}

func parseMarkdownWithFrontmatter(content, repoPath, filename string) (Post, error) {
	var post Post

	if !strings.HasPrefix(content, "---") {
		post.Title = extractTitleFromMarkdown(content)
		post.Content = content
		post.Date = time.Now()

		if repoPath != "" && filename != "" {
			if gitTime := getGitLastModified(repoPath, filename); !gitTime.IsZero() {
				post.LastModified = gitTime
			}
		}

		return post, nil
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return post, fmt.Errorf("invalid frontmatter format")
	}

	frontmatterStr := parts[1]
	markdownContent := strings.TrimSpace(parts[2])

	var frontmatter FrontMatter
	if err := yaml.Unmarshal([]byte(frontmatterStr), &frontmatter); err != nil {
		return post, fmt.Errorf("failed to parse frontmatter: %v", err)
	}

	post.Title = frontmatter.Title
	post.Date = frontmatter.Date
	post.Slug = frontmatter.Slug
	post.Content = markdownContent

	if frontmatter.LastModified != nil {
		post.LastModified = *frontmatter.LastModified
	} else if repoPath != "" && filename != "" {
		if gitTime := getGitLastModified(repoPath, filename); !gitTime.IsZero() {
			post.LastModified = gitTime
		}
	}

	if post.LastModified.IsZero() {
		post.LastModified = post.Date
	}

	if post.Title == "" {
		post.Title = extractTitleFromMarkdown(markdownContent)
	}

	return post, nil
}

func extractTitleFromMarkdown(content string) string {
	re := regexp.MustCompile(`^#\s+(.+)$`)
	for line := range strings.SplitSeq(content, "\n") {
		if matches := re.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
			return strings.TrimSpace(matches[1])
		}
	}
	return "Untitled"
}

func getGitLastModified(repoPath, filename string) time.Time {
	if repoPath == "" || filename == "" {
		return time.Time{}
	}

	cmd := exec.Command("git", "log", "-1", "--format=%ct", "--", filename)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		log.Printf("Git: failed to get last modified time for %s: %v", filename, err)
		return time.Time{}
	}

	timestamp := strings.TrimSpace(string(output))
	if timestamp == "" {
		return time.Time{}
	}

	var unixTime int64
	if _, err := fmt.Sscanf(timestamp, "%d", &unixTime); err != nil {
		log.Printf("Git: failed to parse timestamp %s: %v", timestamp, err)
		return time.Time{}
	}

	return time.Unix(unixTime, 0)
}
