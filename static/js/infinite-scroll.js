(function () {
	var sentinel = document.getElementById('load-more');
	if (!sentinel) return;

	var container = document.querySelector('.container');
	if (!container) return;

	var loading = false;

	var observer = new IntersectionObserver(function (entries) {
		if (entries[0].isIntersecting && !loading) {
			loadMore();
		}
	}, { rootMargin: '200px' });

	observer.observe(sentinel);

	function loadMore() {
		var slug = sentinel.getAttribute('data-after');
		var api = sentinel.getAttribute('data-api');
		if (!slug || !api) return;

		loading = true;
		sentinel.className = 'loading';

		var url = api + '?after=' + encodeURIComponent(slug);

		fetch(url)
			.then(function (res) { return res.json(); })
			.then(function (data) {
				data.posts.forEach(function (p) {
					container.insertBefore(buildCard(p), sentinel);
				});

				if (data.has_next) {
					sentinel.setAttribute('data-after', data.next_slug);
					sentinel.className = '';
					loading = false;
					observer.unobserve(sentinel);
					observer.observe(sentinel);
				} else {
					sentinel.remove();
					observer.disconnect();
				}
			})
			.catch(function () {
				sentinel.className = 'error';
				loading = false;
			});
	}

	function buildCard(p) {
		var art = document.createElement('article');
		art.className = 'post-card';

		var h2 = '<h2><a href="/posts/' + esc(p.slug) + '">' + esc(p.title) + '</a></h2>';

		var meta = '<time datetime="' + esc(p.date) + '">' + esc(p.date) + '</time>';
		if (p.category) {
			meta += ' <a href="/categories/' + esc(p.category.slug) + '" class="category">' + esc(p.category.name) + '</a>';
		}

		var tags = '';
		if (p.tags && p.tags.length) {
			tags = p.tags.map(function (t) {
				return '<a href="/tags/' + esc(t.slug) + '" class="tag">' + esc(t.name) + '</a>';
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
