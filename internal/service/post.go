package service

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"

	"blog/internal/model"
	"blog/internal/store"
)

type PostService struct {
	store *store.Store
	md    goldmark.Markdown
}

func New(s *store.Store) *PostService {
	return &PostService{
		store: s,
		md:    goldmark.New(),
	}
}

func (s *PostService) GetAllPosts() ([]model.Post, error) {
	posts, err := s.store.GetAllPosts()
	if err != nil {
		return nil, err
	}
	for i := range posts {
		posts[i].Summary = s.generateSummary(posts[i].Markdown)
	}
	return posts, nil
}

func (s *PostService) GetPostBySlug(slug string) (*model.Post, error) {
	post, err := s.store.GetPostBySlug(slug)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (s *PostService) RenderMarkdown(input string) (string, error) {
	var buf bytes.Buffer
	if err := s.md.Convert([]byte(input), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *PostService) CreatePost(p *model.Post) error {
	return s.store.CreatePost(p)
}

func (s *PostService) UpdatePost(p *model.Post) error {
	return s.store.UpdatePost(p)
}

func (s *PostService) DeletePost(slug string) error {
	return s.store.DeletePost(slug)
}

func (s *PostService) generateSummary(markdown string) string {
	clean := stripMarkdownSyntax(markdown)
	runes := []rune(clean)
	if len(runes) > 200 {
		return string(runes[:200]) + "..."
	}
	return clean
}

func stripMarkdownSyntax(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "#", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, ">", "")
	s = strings.ReplaceAll(s, "!", "")
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}
