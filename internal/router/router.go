package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"blog/internal/handler/admin"
	"blog/internal/handler/auth"
	"blog/internal/handler/blog"
	mw "blog/internal/middleware"
)

type Deps struct {
	SSR          *blog.SSRHandler
	Auth         *auth.Handler
	Admin        *admin.Handler
	Static       http.Handler
}

func New(deps Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(mw.ForwardedProto)
	r.Use(chimw.Recoverer)
	r.Use(chimw.ClientIPFromRemoteAddr)
	r.Use(mw.CORS)

	r.Handle("/static/*", deps.Static)

	r.Get("/", deps.SSR.Home)
	r.Get("/posts/{slug}", deps.SSR.Post)
	r.Get("/categories/{slug}", deps.SSR.Category)
	r.Get("/tags/{slug}", deps.SSR.Tag)

	r.Get("/api/posts", deps.SSR.APIPosts)
	r.Get("/api/categories", deps.SSR.APISearchCategories)
	r.Get("/api/tags", deps.SSR.APISearchTags)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/admin", func(r chi.Router) {
		r.Post("/login", deps.Auth.Login)

		r.Group(func(r chi.Router) {
			r.Use(deps.Auth.AuthMiddleware)

			r.Get("/me", deps.Auth.Me)
			r.Delete("/logout", deps.Auth.Logout)

			r.Get("/categories", deps.Admin.ListCategories)
			r.Post("/categories", deps.Admin.CreateCategory)
			r.Get("/categories/{id}", deps.Admin.GetCategory)
			r.Put("/categories/{id}", deps.Admin.UpdateCategory)
			r.Delete("/categories/{id}", deps.Admin.DeleteCategory)

			r.Get("/tags", deps.Admin.ListTags)
			r.Post("/tags", deps.Admin.CreateTag)
			r.Get("/tags/{id}", deps.Admin.GetTag)
			r.Put("/tags/{id}", deps.Admin.UpdateTag)
			r.Delete("/tags/{id}", deps.Admin.DeleteTag)

			r.Get("/posts", deps.Admin.ListPosts)
			r.Post("/posts", deps.Admin.CreatePost)
			r.Get("/posts/{id}", deps.Admin.GetPost)
			r.Put("/posts/{id}", deps.Admin.UpdatePost)
			r.Delete("/posts/{id}", deps.Admin.DeletePost)
		})
	})

	r.NotFound(deps.SSR.NotFound)

	return r
}
