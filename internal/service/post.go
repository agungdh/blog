package service

import (
	"bytes"
	"context"
	"strings"

	"github.com/yuin/goldmark"

	"blog/internal/model"
	"blog/internal/store"
)

type PostService struct {
	store *store.Store
	md    goldmark.Markdown
}

type CursorPage struct {
	Posts    []model.Post
	HasNext  bool
	NextSlug string
}

func New(s *store.Store) *PostService {
	return &PostService{
		store: s,
		md:    goldmark.New(),
	}
}

func (s *PostService) toCursorPage(posts []model.Post, limit int) *CursorPage {
	hasNext := len(posts) > limit
	if hasNext {
		posts = posts[:limit]
	}
	for i := range posts {
		posts[i].Summary = s.generateSummary(posts[i].Markdown)
	}
	var nextSlug string
	if hasNext && len(posts) > 0 {
		nextSlug = posts[len(posts)-1].Slug
	}
	return &CursorPage{Posts: posts, HasNext: hasNext, NextSlug: nextSlug}
}

func (s *PostService) GetCursorPosts(ctx context.Context, after string, limit int) (*CursorPage, error) {
	var posts []model.Post
	var err error

	if after != "" {
		posts, err = s.store.GetPostsAfter(ctx, after, limit+1)
	} else {
		posts, err = s.store.GetLatestPosts(ctx, limit+1)
	}
	if err != nil {
		return nil, err
	}

	return s.toCursorPage(posts, limit), nil
}

func (s *PostService) GetCursorPostsFiltered(ctx context.Context, after string, limit int, filter store.FilterParams) (*CursorPage, error) {
	posts, err := s.store.GetFilteredPosts(ctx, filter, limit+1, after)
	if err != nil {
		return nil, err
	}

	return s.toCursorPage(posts, limit), nil
}

func (s *PostService) GetPostBySlug(ctx context.Context, slug string) (*model.Post, error) {
	return s.store.GetPostBySlug(ctx, slug)
}

func (s *PostService) GetLatestPosts(ctx context.Context, limit int) ([]model.Post, error) {
	return s.store.GetLatestPosts(ctx, limit)
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

func (s *PostService) GetCursorPostsByCategorySlug(ctx context.Context, catSlug, after string, limit int) (*CursorPage, error) {
	var posts []model.Post
	var err error

	if after != "" {
		posts, err = s.store.GetPostsByCategorySlugAfter(ctx, catSlug, after, limit+1)
	} else {
		posts, err = s.store.GetLatestPostsByCategorySlug(ctx, catSlug, limit+1)
	}
	if err != nil {
		return nil, err
	}

	return s.toCursorPage(posts, limit), nil
}

func (s *PostService) GetCursorPostsByTagSlug(ctx context.Context, tagSlug, after string, limit int) (*CursorPage, error) {
	var posts []model.Post
	var err error

	if after != "" {
		posts, err = s.store.GetPostsByTagSlugAfter(ctx, tagSlug, after, limit+1)
	} else {
		posts, err = s.store.GetLatestPostsByTagSlug(ctx, tagSlug, limit+1)
	}
	if err != nil {
		return nil, err
	}

	return s.toCursorPage(posts, limit), nil
}

func (s *PostService) GetCategoryBySlug(ctx context.Context, slug string) (*model.Category, error) {
	return s.store.GetCategoryBySlug(ctx, slug)
}

func (s *PostService) SearchCategories(ctx context.Context, query string, limit int) ([]model.Category, error) {
	return s.store.SearchCategories(ctx, query, limit)
}

func (s *PostService) GetCategoriesBySlugs(ctx context.Context, slugs []string) ([]model.Category, error) {
	return s.store.GetCategoriesBySlugs(ctx, slugs)
}

func (s *PostService) GetTagBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	return s.store.GetTagBySlug(ctx, slug)
}

func (s *PostService) SearchTags(ctx context.Context, query string, limit int) ([]model.Tag, error) {
	return s.store.SearchTags(ctx, query, limit)
}

func (s *PostService) GetTagsBySlugs(ctx context.Context, slugs []string) ([]model.Tag, error) {
	return s.store.GetTagsBySlugs(ctx, slugs)
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
