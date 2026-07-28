(function () {
    var sentinel = document.getElementById('load-more');
    if (!sentinel) return;

    var container = sentinel.parentNode;
    if (!container) return;

    var loading = false;

    var observer = new IntersectionObserver(function (entries) {
        if (entries[0].isIntersecting && !loading) {
            loadMore();
        }
    }, {rootMargin: '200px'});

    observer.observe(sentinel);

    function loadMore() {
        var slug = sentinel.getAttribute('data-after');
        var api = sentinel.getAttribute('data-api');
        var filter = sentinel.getAttribute('data-filter') || '';
        if (!slug || !api) return;

        loading = true;
        sentinel.className = 'loading';

        var url = api + '?after=' + encodeURIComponent(slug) + filter;

        fetch(url)
            .then(function (res) {
                if (!res.ok) throw new Error('HTTP ' + res.status);
                return res.json();
            })
            .then(function (data) {
                if (!data.posts) return;

                var frag = document.createDocumentFragment();
                data.posts.forEach(function (p) {
                    frag.appendChild(buildCard(p));
                });

                sentinel.before(frag);

                if (data.has_next) {
                    sentinel.setAttribute('data-after', data.next_slug);
                } else {
                    sentinel.remove();
                    observer.disconnect();
                }
            })
            .catch(function (err) {
                console.error(err);
                sentinel.className = 'error';
            })
            .finally(function () {
                sentinel.classList.remove('loading');
                loading = false;
                observer.unobserve(sentinel);
                if (sentinel.parentNode) observer.observe(sentinel);
            });
    }

    function buildCard(p) {
        var art = document.createElement('article');
        art.className = 'post-card';

        var h2 = '<h2><a href="/posts/' + esc(p.slug) + '">' + esc(p.title) + '</a></h2>';

        var meta = '<time datetime="' + esc(p.date) + '">' + esc(p.date) + '</time>';
        if (p.category) {
            meta += ' <a href="/?category=' + esc(p.category.slug) + '" class="category">' + esc(p.category.name) + '</a>';
        }

        var tags = '';
        if (p.tags && p.tags.length) {
            tags = p.tags.map(function (t) {
                return '<a href="/?tags=' + esc(t.slug) + '" class="tag">' + esc(t.name) + '</a>';
            }).join('');
        }

        art.innerHTML =
            h2 +
            '<div class="post-meta">' + meta + '</div>' +
            '<p class="post-summary">' + esc(p.summary) + '</p>' +
            (tags ? '<div class="post-tags">' + tags + '</div>' : '');

        return art;
    }

    function esc(s) {
        return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
    }
})();
