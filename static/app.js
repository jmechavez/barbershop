// DRFT Barbershop — small motion layer.
// No frameworks. No dependencies. Fails gracefully.

// 1. Reveal elements on scroll.
(function () {
	const els = document.querySelectorAll('[data-reveal]');
	if (!els.length || !('IntersectionObserver' in window)) return;

	const io = new IntersectionObserver((entries) => {
		entries.forEach((entry) => {
			if (entry.isIntersecting) {
				entry.target.classList.add('revealed');
				io.unobserve(entry.target);
			}
		});
	}, { threshold: 0.12, rootMargin: '0px 0px -60px 0px' });

	els.forEach((el) => io.observe(el));
})();

// 2. Row selection state (visual only).
(function () {
	const rows = document.querySelectorAll('.svc-row');
	rows.forEach((row) => {
		const input = row.querySelector('input[type="checkbox"]');
		if (!input) return;
		row.addEventListener('click', () => {
			setTimeout(() => row.classList.toggle('selected', input.checked), 0);
		});
		if (input.checked) row.classList.add('selected');
	});
})();

// 3. Smooth scroll for nav links.
(function () {
	document.querySelectorAll('a[href^="#"]').forEach((a) => {
		a.addEventListener('click', (e) => {
			const id = a.getAttribute('href').slice(1);
			const el = document.getElementById(id);
			if (!el) return;
			e.preventDefault();
			el.scrollIntoView({ behavior: 'smooth', block: 'start' });
		});
	});
})();

// 4. Close mobile nav when a link is clicked.
(function () {
	const nav = document.getElementById('topnav');
	if (!nav) return;
	nav.querySelectorAll('.nav-links a').forEach((a) => {
		a.addEventListener('click', () => nav.classList.remove('open'));
	});
})();

// 5. Active nav link on scroll.
(function () {
	const sections = document.querySelectorAll('section[id]');
	const links = document.querySelectorAll('.nav-links a[href^="#"]');
	if (!sections.length || !links.length) return;

	const linkFor = {};
	links.forEach((a) => { linkFor[a.getAttribute('href').slice(1)] = a; });

	const io = new IntersectionObserver((entries) => {
		entries.forEach((entry) => {
			const a = linkFor[entry.target.id];
			if (!a) return;
			if (entry.isIntersecting) {
				links.forEach((l) => l.classList.remove('active'));
				a.classList.add('active');
			}
		});
	}, { threshold: 0.4 });

	sections.forEach((s) => io.observe(s));
})();

// 6. Scroll progress bar.
(function () {
	const bar = document.createElement('div');
	bar.className = 'scroll-progress';
	document.body.appendChild(bar);

	function update() {
		const h = document.documentElement;
		const scrolled = h.scrollTop / (h.scrollHeight - h.clientHeight);
		bar.style.transform = 'scaleX(' + Math.min(1, Math.max(0, scrolled)) + ')';
	}
	window.addEventListener('scroll', update, { passive: true });
	update();
})();

// 7. Preserve checked services in sessionStorage.
(function () {
	const form = document.getElementById('pick');
	if (!form) return;

	const KEY = 'drft_picked';

	// On load: restore any previously checked boxes.
	try {
		const saved = JSON.parse(sessionStorage.getItem(KEY) || '[]');
		form.querySelectorAll('input[name="service"]').forEach((input) => {
			if (saved.includes(input.value)) {
				input.checked = true;
				const row = input.closest('.svc-row');
				if (row) row.classList.add('selected');
			}
		});
	} catch (e) { /* ignore */ }

	// On change: save the current checked set.
	form.addEventListener('change', () => {
		const picked = Array.from(
			form.querySelectorAll('input[name="service"]:checked')
		).map((i) => i.value);
		try {
			sessionStorage.setItem(KEY, JSON.stringify(picked));
		} catch (e) { /* ignore */ }
	});
})();

// 8. Animate the total when it changes (htmx swap).
(function () {
	const amount = document.getElementById('total-amount');
	if (!amount) return;

	const mo = new MutationObserver(() => {
		amount.style.transform = 'scale(1.08)';
		setTimeout(() => { amount.style.transform = ''; }, 180);
	});
	mo.observe(amount, { childList: true, characterData: true, subtree: true });
})();