document.addEventListener('alpine:init', () => {
    Alpine.data('categorySearch', () => ({
        slug: '',
        results: [],
        show: false,
        get $input() { return this.$el.querySelector('input[type="text"]') },
        init() {
            this.slug = this.$el.dataset.catSlug || '';
        },
        search() {
            var q = this.$input ? this.$input.value.trim() : '';
            if (!q) { this.results = []; this.show = false; return; }
            fetch('/api/categories?q=' + encodeURIComponent(q))
                .then(r => r.json())
                .then(data => { this.results = data || []; this.show = (data && data.length > 0); });
        },
        select(item) {
            if (this.$input) this.$input.value = item.name;
            this.slug = item.slug;
            this.results = [];
            this.show = false;
        },
        blur() { var self = this; setTimeout(() => { self.results = []; self.show = false; }, 150); }
    }));

    Alpine.data('tagChips', () => ({
        tags: [],
        results: [],
        show: false,
        get $input() { return this.$el.querySelector('input[type="text"]') },
        init() {
            var el = document.getElementById('init-tags-data');
            try { this.tags = el ? JSON.parse(el.textContent) : []; } catch(e) { this.tags = []; }
        },
        search() {
            var q = this.$input ? this.$input.value.trim() : '';
            if (!q) { this.results = []; this.show = false; return; }
            fetch('/api/tags?q=' + encodeURIComponent(q))
                .then(r => r.json())
                .then(data => { this.results = data || []; this.show = (data && data.length > 0); });
        },
        add(item) {
            if (!this.tags.find(t => t.slug === item.slug)) this.tags.push(item);
            if (this.$input) this.$input.value = '';
            this.results = [];
            this.show = false;
        },
        remove(index) { this.tags.splice(index, 1); },
        blur() { var self = this; setTimeout(() => { self.results = []; self.show = false; }, 150); }
    }));

    Alpine.data('infiniteScroll', () => ({
        loading: false,
        error: false,
        done: false,
        init() {
            var self = this;
            window.addEventListener('scroll', () => self.onScroll(), {passive: true});
            this.$nextTick(() => this.onScroll());
        },
        onScroll() {
            if (this.loading || this.done) return;
            var rect = this.$el.getBoundingClientRect();
            if (rect.top < window.innerHeight + 400) this.load();
        },
        load() {
            if (this.loading || this.done) return;
            var el = this.$el;
            var api = el.dataset.api, after = el.dataset.after, filter = el.dataset.filter || '';
            if (!api || !after) { this.done = true; return; }
            this.loading = true; this.error = false;
            var self = this;
            fetch(api + '?after=' + encodeURIComponent(after) + filter)
                .then(r => { if (!r.ok) throw new Error(); return r.json(); })
                .then(data => {
                    if (!data.posts || data.posts.length === 0) { this.done = true; return; }
                    el.insertAdjacentHTML('beforebegin', data.posts.map(p => buildCard(p)).join(''));
                    if (data.has_next) {
                        el.dataset.after = data.next_slug;
                        self.$nextTick(() => self.onScroll());
                    } else {
                        this.done = true;
                    }
                })
                .catch(() => { this.error = true; })
                .finally(() => { this.loading = false; });
        }
    }));
});

function esc(s) {
    return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function buildCard(p) {
    var cat = p.category
        ? '<a class="tag is-link is-light is-small" href="/?category=' + esc(p.category.slug) + '">' + esc(p.category.name) + '</a>'
        : '';
    var meta = '<div class="is-flex is-align-items-center mb-3" style="gap:0.5rem">'
        + cat
        + '<time class="has-text-grey is-size-7" datetime="' + esc(p.date) + '">' + esc(p.date) + '</time>'
        + '</div>';
    var tags = '';
    if (p.tags && p.tags.length) {
        tags = '<div class="tags mt-3">' + p.tags.map(function(t) {
            return '<a class="tag is-light" href="/?tags=' + esc(t.slug) + '">' + esc(t.name) + '</a>';
        }).join('') + '</div>';
    }
    return '<div class="card mb-5"><div class="card-content">'
        + meta
        + '<h2 class="title is-4 mb-2"><a href="/posts/' + esc(p.slug) + '">' + esc(p.title) + '</a></h2>'
        + '<p class="has-text-grey-dark" style="line-height:1.7">' + esc(p.summary) + '</p>'
        + tags
        + '</div></div>';
}
