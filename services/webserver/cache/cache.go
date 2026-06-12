package cache

import (
	"fmt"
	"log"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"personal-website/internal"
)

const (
	postCacheTTL       = 5 * time.Minute
	absentSlugTTL      = time.Minute
	absentSlugCapacity = 1024
)

type PostCache struct {
	posts           map[string]internal.Post
	allPosts        []internal.Post
	recentlyAbsent  map[string]time.Time
	mutex           sync.RWMutex
	lastFetch       time.Time
	refreshInFlight atomic.Bool
}

var Cache = &PostCache{
	posts:          make(map[string]internal.Post),
	allPosts:       []internal.Post{},
	recentlyAbsent: make(map[string]time.Time),
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
	c.recentlyAbsent = make(map[string]time.Time)

	c.lastFetch = time.Now()
	log.Println("Cache: updated successfully")
	return nil
}

func (c *PostCache) refreshIfStale(lastFetch time.Time) {
	if time.Since(lastFetch) <= postCacheTTL {
		return
	}
	if !c.refreshInFlight.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer c.refreshInFlight.Store(false)
		if err := c.updateCache(); err != nil {
			log.Printf("Cache: background refresh failed: %v", err)
		}
	}()
}

func (c *PostCache) GetPosts() ([]internal.Post, error) {
	c.mutex.RLock()
	posts := slices.Clone(c.allPosts)
	lastFetch := c.lastFetch
	c.mutex.RUnlock()

	c.refreshIfStale(lastFetch)

	if len(posts) == 0 && lastFetch.IsZero() {
		return nil, fmt.Errorf("cache not initialized")
	}

	return posts, nil
}

func (c *PostCache) GetPost(slug string) (internal.Post, error) {
	c.mutex.RLock()
	post, found := c.posts[slug]
	lastFetch := c.lastFetch
	c.mutex.RUnlock()

	c.refreshIfStale(lastFetch)

	if !found {
		return internal.Post{}, fmt.Errorf("post %q not in cache", slug)
	}

	return post, nil
}

func (c *PostCache) MarkAbsent(slug string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(c.recentlyAbsent) >= absentSlugCapacity {
		c.recentlyAbsent = make(map[string]time.Time, absentSlugCapacity)
	}
	c.recentlyAbsent[slug] = time.Now()
}

func (c *PostCache) IsRecentlyAbsent(slug string) bool {
	c.mutex.RLock()
	markedAt, found := c.recentlyAbsent[slug]
	c.mutex.RUnlock()
	return found && time.Since(markedAt) < absentSlugTTL
}

func InitCache() {
	if err := Cache.updateCache(); err != nil {
		log.Printf("Cache: failed to initialize cache: %v", err)
	}
}
