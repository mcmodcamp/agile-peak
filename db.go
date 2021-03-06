package main

import (
	"net/url"
	"strconv"

	"github.com/satori/go.uuid"
	"gopkg.in/redis.v4"
)

type DB interface {
	Clear(page string) error
	Get(uuid uuid.UUID) (*Post, error)
	List(page string) ([]uuid.UUID, error)
	ListPages() ([]string, error)
	Post(page string, post *Post) (uuid.UUID, error)
}

type Post struct {
	Name string
	ID   int
	Text string
}

type db struct {
	client *redis.Client
}

func (db *db) Clear(page string) error {
	return db.client.Del("/" + page).Err()
}

func (db *db) Get(uuid uuid.UUID) (*Post, error) {
	m, err := db.client.HGetAll(uuid.String()).Result()
	if err != nil {
		return nil, err
	}

	id, err := strconv.Atoi(m["id"])
	if err != nil {
		return nil, err
	}

	return &Post{
		Name: m["name"],
		ID:   id,
		Text: m["text"],
	}, nil
}

func (db *db) List(page string) ([]uuid.UUID, error) {
	strs, err := db.client.LRange("/"+page, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	uuids := make([]uuid.UUID, len(strs))
	for i, str := range strs {
		uuids[i], err = uuid.FromString(str)
		if err != nil {
			return nil, err
		}
	}
	return uuids, err
}

func (db *db) ListPages() ([]string, error) {
	return db.client.SMembers("pages").Result()
}

func (db *db) Post(page string, post *Post) (uuid.UUID, error) {
	uuid := uuid.NewV4()
	if err := db.client.HSet(uuid.String(), "name", post.Name).Err(); err != nil {
		return uuid, err
	}
	if err := db.client.HSet(uuid.String(), "id", strconv.Itoa(post.ID)).Err(); err != nil {
		return uuid, err
	}
	if err := db.client.HSet(uuid.String(), "text", post.Text).Err(); err != nil {
		return uuid, err
	}

	if err := db.client.RPush("/"+page, uuid.String()).Err(); err != nil {
		return uuid, err
	}

	return uuid, db.client.SAdd("pages", page).Err()
}

func ConnectDB(addr string) (DB, error) {
	url, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	var password string
	if pw, ok := url.User.Password(); ok {
		password = pw
	}
	client := redis.NewClient(&redis.Options{
		Addr:     url.Host,
		Password: password,
	})
	return &db{client}, nil
}
