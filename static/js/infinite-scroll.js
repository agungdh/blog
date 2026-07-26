(function () {
	var container = document.querySelector('.container');
	if (!container) return;

	var sentinel = document.getElementById('load-more');
	if (!sentinel) return;

	var loading = false;

	var observer = new IntersectionObserver(function (entries) {
		if (entries[0].isIntersecting && !loading) {
			loadMore();
		}
	}, { rootMargin: '200px' });

	observer.observe(sentinel);

	function loadMore() {
		var slug = sentinel.getAttribute('data-after');
		if (!slug) return;

		loading = true;
		sentinel.className = 'loading';

		var url = window.location.pathname + '?after=' + encodeURIComponent(slug);

		fetch(url)
			.then(function (res) { return res.text(); })
			.then(function (html) {
				var parser = new DOMParser();
				var doc = parser.parseFromString(html, 'text/html');

				var cards = doc.querySelectorAll('.post-card');
				cards.forEach(function (card) {
					container.insertBefore(card, sentinel);
				});

				var next = doc.getElementById('load-more');
				if (next && next.getAttribute('data-after')) {
					sentinel.setAttribute('data-after', next.getAttribute('data-after'));
					sentinel.className = '';
				} else {
					sentinel.remove();
					observer.disconnect();
					return;
				}

				loading = false;
				observer.unobserve(sentinel);
				observer.observe(sentinel);
			})
			.catch(function () {
				sentinel.className = 'error';
				loading = false;
			});
	}
})();
