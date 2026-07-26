package service

import (
	"bytes"
	"context"
	"math"
	"strings"

	"github.com/yuin/goldmark"

	"blog/internal/model"
	"blog/internal/store"
)

type PostService struct {
	store *store.Store
	md    goldmark.Markdown
}

type PagedPosts struct {
	Posts      []model.Post
	Total      int
	Page       int
	PerPage    int
	TotalPages int
	HasNext    bool
	HasPrev    bool
	PrevPage   int
	NextPage   int
}

func New(s *store.Store) *PostService {
	return &PostService{
		store: s,
		md:    goldmark.New(),
	}
}

func (s *PostService) GetPagedPosts(ctx context.Context, page, perPage int) (*PagedPosts, error) {
	if page < 1 {
		page = 1
	}

	total, err := s.store.CountPosts(ctx)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages > 0 && page > totalPages {
		page = totalPages
	}
	offset := (page - 1) * perPage

	posts, err := s.store.GetPosts(ctx, offset, perPage)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		posts[i].Summary = s.generateSummary(posts[i].Markdown)
	}

	return &PagedPosts{
		Posts:      posts,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
		PrevPage:   page - 1,
		NextPage:   page + 1,
	}, nil
}

func (s *PostService) GetPostBySlug(ctx context.Context, slug string) (*model.Post, error) {
	return s.store.GetPostBySlug(ctx, slug)
}

func (s *PostService) RenderMarkdown(input string) (string, error) {
	var buf bytes.Buffer
	if err := s.md.Convert([]byte(input), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *PostService) CreatePost(ctx context.Context, p *model.Post) error {
	return s.store.CreatePost(ctx, p)
}

func (s *PostService) UpdatePost(ctx context.Context, p *model.Post) error {
	return s.store.UpdatePost(ctx, p)
}

func (s *PostService) DeletePost(ctx context.Context, slug string) error {
	return s.store.DeletePost(ctx, slug)
}

func (s *PostService) GetPagedPostsByCategorySlug(ctx context.Context, slug string, page, perPage int) (*PagedPosts, error) {
	if page < 1 {
		page = 1
	}

	total, err := s.store.CountPostsByCategorySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages > 0 && page > totalPages {
		page = totalPages
	}
	offset := (page - 1) * perPage

	posts, err := s.store.GetPostsByCategorySlug(ctx, slug, offset, perPage)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		posts[i].Summary = s.generateSummary(posts[i].Markdown)
	}

	return &PagedPosts{
		Posts:      posts,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}, nil
}

func (s *PostService) GetPagedPostsByTagSlug(ctx context.Context, slug string, page, perPage int) (*PagedPosts, error) {
	if page < 1 {
		page = 1
	}

	total, err := s.store.CountPostsByTagSlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages > 0 && page > totalPages {
		page = totalPages
	}
	offset := (page - 1) * perPage

	posts, err := s.store.GetPostsByTagSlug(ctx, slug, offset, perPage)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		posts[i].Summary = s.generateSummary(posts[i].Markdown)
	}

	return &PagedPosts{
		Posts:      posts,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}, nil
}

func (s *PostService) GetCategoryBySlug(ctx context.Context, slug string) (*model.Category, error) {
	return s.store.GetCategoryBySlug(ctx, slug)
}

func (s *PostService) GetTagBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	return s.store.GetTagBySlug(ctx, slug)
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
