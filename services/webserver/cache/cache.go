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

	db, err := internal.CreateConnection()
	defer db.Close()
	if err != nil {
		return err
	}

	posts, err := internal.GetPosts(db)
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

	if time.Since(c.lastFetch) > 5*time.Minute {
		go c.updateCache() // Update cache asynchronously
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
