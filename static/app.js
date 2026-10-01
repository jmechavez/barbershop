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