(function () {
	var CAT_API = '/api/categories';
	var TAG_API = '/api/tags';

	initCategorySearch();
	initTagsSearch();

	function initCategorySearch() {
		var wrap = document.getElementById('category-search-wrapper');
		if (!wrap) return;
		var input = document.getElementById('category-search');
		var hidden = wrap.querySelector('input[name="category"]');
		var dropdown = wrap.querySelector('.search-dropdown');
		if (!input || !dropdown) return;

		var timer;

		input.addEventListener('input', function () {
			clearTimeout(timer);
			if (hidden) hidden.value = '';
			timer = setTimeout(function () {
				fetchResults(CAT_API, input.value.trim(), dropdown, function (item) {
					input.value = item.name;
					if (hidden) hidden.value = item.slug;
					dropdown.style.display = 'none';
				});
			}, 300);
		});

		input.addEventListener('focus', function () {
			if (dropdown.children.length > 0) dropdown.style.display = 'block';
		});

		document.addEventListener('click', function (e) {
			if (!wrap.contains(e.target)) dropdown.style.display = 'none';
		});
	}

	function initTagsSearch() {
		var wrap = document.getElementById('tags-search-wrapper');
		if (!wrap) return;
		var input = document.getElementById('tags-search');
		var chips = document.getElementById('tag-chips');
		var dropdown = wrap.querySelector('.search-dropdown');
		var hiddenC = document.getElementById('tags-hidden');
		if (!input || !dropdown || !chips) return;

		var timer;

		input.addEventListener('input', function () {
			clearTimeout(timer);
			timer = setTimeout(function () {
				fetchResults(TAG_API, input.value.trim(), dropdown, function (item) {
					addTagChip(chips, hiddenC, item);
					input.value = '';
					dropdown.style.display = 'none';
				});
			}, 300);
		});

		input.addEventListener('focus', function () {
			if (dropdown.children.length > 0) dropdown.style.display = 'block';
		});

		document.addEventListener('click', function (e) {
			if (!wrap.contains(e.target)) dropdown.style.display = 'none';
		});

		chips.addEventListener('click', function (e) {
			if (e.target.classList.contains('tag-chip-remove')) {
				var chip = e.target.parentNode;
				var slug = chip.getAttribute('data-slug');
				chip.remove();
				var cb = hiddenC.querySelector('input[value="' + escAttr(slug) + '"]');
				if (cb) cb.remove();
			}
		});
	}

	function addTagChip(chips, hiddenC, item) {
		if (chips.querySelector('.tag-chip[data-slug="' + escAttr(item.slug) + '"]')) return;

		var chip = document.createElement('span');
		chip.className = 'tag-chip';
		chip.setAttribute('data-slug', item.slug);
		chip.innerHTML = escHtml(item.name) + '<button type="button" class="tag-chip-remove">&times;</button>';
		chips.appendChild(chip);

		var h = document.createElement('input');
		h.type = 'checkbox';
		h.name = 'tags';
		h.value = item.slug;
		h.checked = true;
		h.style.display = 'none';
		hiddenC.appendChild(h);
	}

	function fetchResults(api, query, dropdown, onSelect) {
		fetch(api + '?q=' + encodeURIComponent(query))
			.then(function (res) { return res.json(); })
			.then(function (data) {
				dropdown.innerHTML = '';
				if (!data || !data.length) {
					dropdown.style.display = 'none';
					return;
				}
				data.forEach(function (item) {
					var div = document.createElement('div');
					div.className = 'search-dropdown-item';
					div.textContent = item.name;
					div.addEventListener('mousedown', function (e) {
						e.preventDefault();
						onSelect(item);
					});
					dropdown.appendChild(div);
				});
				dropdown.style.display = 'block';
			})
			.catch(function () {
				dropdown.style.display = 'none';
			});
	}

	function escHtml(s) {
		return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
	}

	function escAttr(s) {
		return String(s).replace(/"/g, '\\"');
	}
})();
