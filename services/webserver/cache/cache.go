package cache

import (
	"fmt"
	"log"
	"sync"
	"time"

	"personal-website/internal"
)

type PostCache struct {
	posts     map[string]internal.Post
	allPosts  []internal.Post
	mutex     sync.RWMutex
	lastFetch time.Time
}

var Cache = &PostCache{
	posts:    make(map[string]internal.Post),
	allPosts: []internal.Post{},
}

func (c *PostCache) updateCache() error {
	log.Println("Cache: Updating...")

	posts, err := internal.GetPostsFromGitHub()
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.allPosts = posts
	c.posts = make(map[string]internal.Post)
	for _, post := range posts {
		c.posts[post.Slug] = post
	}

	c.lastFetch = time.Now()
	log.Println("Cache: updated successfully")
	return nil
}

func (c *PostCache) GetPosts() ([]internal.Post, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	log.Printf("Cache: GetPosts called, have %d posts, last fetch: %v", len(c.allPosts), c.lastFetch)

	if time.Since(c.lastFetch) > 5*time.Minute {
		log.Printf("Cache: Cache expired, triggering async update")
		go c.updateCache() // Update cache asynchronously
	}

	if len(c.allPosts) == 0 && c.lastFetch.IsZero() {
		log.Printf("Cache: No posts and never fetched, this is likely an initialization issue")
		return nil, fmt.Errorf("cache not initialized")
	}

	return c.allPosts, nil
}

func (c *PostCache) GetPost(slug string) (internal.Post, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if time.Since(c.lastFetch) > 5*time.Minute {
		go c.updateCache()
	}

	post, ok := c.posts[slug]
	if !ok {
		return internal.Post{}, fmt.Errorf("Cache: post not found")
	}

	return post, nil
}

func InitCache() {
	err := Cache.updateCache()
	if err != nil {
		log.Printf("Cache: failed to initialize cache: %v", err)
	}
}
