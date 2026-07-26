package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"blog/internal/model"
	"blog/internal/store"
)

type PostGenerator struct {
	ai    *AIClient
	store *store.Store
}

func NewPostGenerator(ai *AIClient, store *store.Store) *PostGenerator {
	return &PostGenerator{ai: ai, store: store}
}

type generatedPost struct {
	Title    string           `json:"title"`
	Slug     string           `json:"slug"`
	Content  string           `json:"content"`
	Excerpt  string           `json:"excerpt"`
	Category generatedCategory `json:"category"`
	Tags     []generatedTag   `json:"tags"`
}

type generatedCategory struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type generatedTag struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (g *PostGenerator) Generate(ctx context.Context) {
	log.Printf("generator: starting post generation")

	existing, err := g.store.GetAllPostTitles(ctx)
	if err != nil {
		log.Printf("generator: failed to get existing posts: %v", err)
		return
	}

	systemPrompt := `Kamu adalah penulis blog teknologi profesional. Hasilkan SATU artikel blog teknologi baru dalam bahasa Indonesia.
Format response HARUS JSON saja, tanpa markdown wrapper, tanpa penjelasan. Format:

{
  "title": "Judul artikel yang SEO-friendly dan menarik",
  "slug": "judul-artikel-dalam-bahasa-inggris-kebab-case",
  "content": "Konten artikel dalam format Markdown. Gunakan ## untuk heading, **bold**, daftar, dan code blocks dengan triple backtick. Buat artikel informatif, 500-1000 kata, dengan contoh kode yang relevan.",
  "excerpt": "Ringkasan 1-3 kalimat dalam bahasa Indonesia",
  "category": {"name": "Nama Kategori", "slug": "nama-kategori"},
  "tags": [{"name": "Nama Tag", "slug": "nama-tag"}, ...]
}

Topik: teknologi modern (web development, cloud, DevOps, programming, AI/ML, cybersecurity, mobile dev, data science).
Slug harus unik, bahasa Inggris lowercase, kebab-case.
Gunakan minimal 2 tag.
`

	userPrompt := "Post yang SUDAH ADA (JANGAN buat judul yang sama):\n"
	for _, t := range existing {
		userPrompt += "- " + t + "\n"
	}
	userPrompt += "\nBuat 1 artikel baru yang berbeda dari semua judul di atas."

	raw, err := g.ai.Generate(systemPrompt, userPrompt)
	if err != nil {
		log.Printf("generator: ai call failed: %v", err)
		return
	}

	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var gp generatedPost
	if err := json.Unmarshal([]byte(raw), &gp); err != nil {
		log.Printf("generator: failed to parse ai response: %v\nRaw: %s", err, raw[:min(len(raw), 500)])
		return
	}

	if gp.Title == "" || gp.Content == "" {
		log.Printf("generator: ai returned empty post")
		return
	}

	category, err := g.upsertCategory(ctx, gp.Category.Name, gp.Category.Slug)
	if err != nil {
		log.Printf("generator: failed to upsert category: %v", err)
		return
	}

	var tags []model.Tag
	for _, t := range gp.Tags {
		tag, err := g.upsertTag(ctx, t.Name, t.Slug)
		if err != nil {
			log.Printf("generator: failed to upsert tag %q: %v", t.Slug, err)
			continue
		}
		tags = append(tags, *tag)
	}

	slug := gp.Slug
	if slug == "" {
		slug = generateSlug(gp.Title)
	}
	slug = g.resolveSlug(ctx, slug)

	post := &model.Post{
		Slug:       slug,
		Title:      gp.Title,
		Date:       time.Now().Format("2006-01-02"),
		Markdown:   gp.Content,
		CategoryID: category.ID,
		Tags:       tags,
	}

	if err := g.store.CreatePost(ctx, post); err != nil {
		log.Printf("generator: failed to create post: %v", err)
		return
	}

	log.Printf("generator: created post %q (slug: %s, category: %s, tags: %d)",
		gp.Title, slug, category.Name, len(tags))
}

func (g *PostGenerator) upsertCategory(ctx context.Context, name, slug string) (*model.Category, error) {
	cat, err := g.store.GetCategoryBySlug(ctx, slug)
	if err == nil {
		return cat, nil
	}
	cat = &model.Category{Name: name, Slug: slug}
	if err := g.store.CreateCategory(ctx, cat); err != nil {
		return nil, fmt.Errorf("create category %q: %w", slug, err)
	}
	return cat, nil
}

func (g *PostGenerator) upsertTag(ctx context.Context, name, slug string) (*model.Tag, error) {
	tag, err := g.store.GetTagBySlug(ctx, slug)
	if err == nil {
		return tag, nil
	}
	tag = &model.Tag{Name: name, Slug: slug}
	if err := g.store.CreateTag(ctx, tag); err != nil {
		return nil, fmt.Errorf("create tag %q: %w", slug, err)
	}
	return tag, nil
}

func (g *PostGenerator) resolveSlug(ctx context.Context, slug string) string {
	baseSlug := slug
	for i := 0; i < 100; i++ {
		_, err := g.store.GetPostBySlug(ctx, slug)
		if err != nil {
			return slug
		}
		suffix := fmt.Sprintf("-%d", rand.Intn(10000))
		slug = baseSlug + suffix
	}
	return baseSlug + "-" + fmt.Sprintf("%d", time.Now().Unix())
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	return strings.Trim(slug, "-")
}
