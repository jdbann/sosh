package store

type MemoryStore struct {
	Posts []Post
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) GetPosts() ([]Post, error) {
	return s.Posts, nil
}

func (s *MemoryStore) AddPost(p Post) error {
	s.Posts = append(s.Posts, p)
	return nil
}
