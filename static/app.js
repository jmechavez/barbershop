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

	form.addEventListener('change', () => {
		const picked = Array.from(
			form.querySelectorAll('input[name="service"]:checked')
		).map((i) => i.value);
		try {
			sessionStorage.setItem(KEY, JSON.stringify(picked));
		} catch (e) { /* ignore */ }
	});
})();

// 8. Public page — animate the total number from old to new (htmx swap).
(function () {
	const el = document.getElementById('total-amount');
	if (!el) return;

	let currentValue = 0;
	let rafId = null;
	let animating = false;

	function parsePesos(str) {
		const cleaned = (str || '').replace(/[^\d-]/g, '');
		const n = parseInt(cleaned, 10);
		return isNaN(n) ? 0 : n;
	}

	function formatCentavos(centavos) {
		const sign = centavos < 0 ? '-' : '';
		const abs = Math.abs(centavos);
		const whole = Math.floor(abs / 100);
		const frac = abs % 100;
		const wholeStr = String(whole).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
		return sign + '₱' + wholeStr + '.' + String(frac).padStart(2, '0');
	}

	currentValue = parsePesos(el.textContent || '0');

	function animateTo(target) {
		if (rafId) cancelAnimationFrame(rafId);
		const start = currentValue;
		const delta = target - start;
		if (delta === 0) return;

		animating = true;
		const duration = 800;
		const t0 = performance.now();

		function tick(now) {
			const t = Math.min(1, (now - t0) / duration);
			const eased = 1 - Math.pow(1 - t, 3);
			const value = Math.round(start + delta * eased);
			el.textContent = formatCentavos(value);
			currentValue = value;
			if (t < 1) {
				rafId = requestAnimationFrame(tick);
			} else {
				rafId = null;
				animating = false;
			}
		}
		rafId = requestAnimationFrame(tick);
	}

	const mo = new MutationObserver(() => {
		if (animating) return; // ignore our own writes
		const target = parsePesos(el.textContent || '0');
		if (target !== currentValue) {
			animateTo(target);
		}
	});
	mo.observe(el, { childList: true, characterData: true, subtree: true });
})();

// 9. Mark body as loaded after page finishes (for skeleton loading).
(function () {
	function markLoaded() {
		if (document.body) document.body.classList.add('loaded');
	}
	if (document.readyState === 'complete') {
		markLoaded();
	} else {
		window.addEventListener('load', markLoaded);
	}
})();

// 10. Custom confirm modal.
(function () {
	const modal = document.getElementById('confirmModal');
	if (!modal) return;

	const messageEl = document.getElementById('confirmMessage');
	const cancelBtn = document.getElementById('confirmCancel');
	const okBtn = document.getElementById('confirmOk');

	let pendingCallback = null;

	function openConfirm(message, callback) {
		messageEl.textContent = message;
		pendingCallback = callback;
		modal.classList.add('open');
		modal.setAttribute('aria-hidden', 'false');
		okBtn.focus();
	}

	function closeConfirm() {
		modal.classList.remove('open');
		modal.setAttribute('aria-hidden', 'true');
		pendingCallback = null;
	}

	cancelBtn.addEventListener('click', closeConfirm);

	okBtn.addEventListener('click', () => {
		const cb = pendingCallback;
		closeConfirm();
		if (cb) cb();
	});

	modal.addEventListener('click', (e) => {
		if (e.target === modal) closeConfirm();
	});

	document.addEventListener('keydown', (e) => {
		if (e.key === 'Escape' && modal.classList.contains('open')) {
			closeConfirm();
		}
	});

	window.drftConfirm = openConfirm;
})();

// 11. Intercept clicks on .js-confirm elements.
(function () {
	document.addEventListener('click', (e) => {
		const target = e.target.closest('.js-confirm');
		if (!target) return;
		if (!window.drftConfirm) return;

		e.preventDefault();
		e.stopPropagation();

		const message = target.getAttribute('data-confirm') || 'Are you sure?';

		window.drftConfirm(message, () => {
			target.classList.remove('js-confirm');
			target.click();
			setTimeout(() => target.classList.add('js-confirm'), 100);
		});
	}, true);
})();

// 12. Counter — live payment calculation.
(function () {
	const form = document.getElementById('counter-form');
	if (!form) return;

	const serviceSelect = document.getElementById('service-select');
	const discountInput = document.getElementById('discount-input');
	const netDueEl = document.getElementById('net-due');
	const totalPaidEl = document.getElementById('total-paid');
	const remainingEl = document.getElementById('remaining');
	const totalsBox = document.getElementById('payment-totals');
	const payInputs = form.querySelectorAll('.pay-input');

	function pesos(centavos) {
		const sign = centavos < 0 ? '-' : '';
		const abs = Math.abs(centavos);
		const whole = Math.floor(abs / 100);
		const frac = abs % 100;
		const wholeStr = String(whole).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
		return sign + '₱' + wholeStr + '.' + String(frac).padStart(2, '0');
	}

	function currentServicePriceCentavos() {
		if (!serviceSelect) return 0;
		const opt = serviceSelect.options[serviceSelect.selectedIndex];
		if (!opt) return 0;
		const p = parseInt(opt.getAttribute('data-price') || '0', 10);
		return isNaN(p) ? 0 : p;
	}

	function currentDiscountCentavos() {
		if (!discountInput) return 0;
		const p = parseInt(discountInput.value || '0', 10);
		return isNaN(p) ? 0 : p * 100;
	}

	function totalPaidCentavos() {
		let total = 0;
		payInputs.forEach((inp) => {
			const p = parseInt(inp.value || '0', 10);
			if (!isNaN(p) && p > 0) total += p * 100;
		});
		return total;
	}

	function update() {
		const price = currentServicePriceCentavos();
		const discount = Math.min(currentDiscountCentavos(), price);
		const net = price - discount;
		const paid = totalPaidCentavos();
		const remaining = net - paid;

		netDueEl.textContent = pesos(net);
		totalPaidEl.textContent = pesos(paid);
		remainingEl.textContent = pesos(remaining);

		payInputs.forEach((inp) => {
			const card = inp.closest('.pay-card');
			const v = parseInt(inp.value || '0', 10);
			card.classList.toggle('has-value', !isNaN(v) && v > 0);
		});

		totalsBox.classList.remove('complete', 'short', 'over');
		if (remaining === 0 && net > 0) {
			totalsBox.classList.add('complete');
		} else if (remaining > 0) {
			totalsBox.classList.add('short');
		} else if (remaining < 0) {
			totalsBox.classList.add('over');
		}
	}

	if (serviceSelect) serviceSelect.addEventListener('change', update);
	if (discountInput) discountInput.addEventListener('input', update);
	payInputs.forEach((inp) => inp.addEventListener('input', update));

	update();
})();

// 13. Counter — auto-fill the datetime field with "now".
(function () {
	const input = document.getElementById('created-at');
	if (!input) return;

	function pad(n) {
		return String(n).padStart(2, '0');
	}

	function setNow() {
		const d = new Date();
		const s = d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate())
		        + 'T' + pad(d.getHours()) + ':' + pad(d.getMinutes());
		input.value = s;
	}

	setNow();
})();

// 14. Counter — quick discount buttons.
(function () {
	const discountInput = document.getElementById('discount-input');
	const serviceSelect = document.getElementById('service-select');
	const quickButtons = document.querySelectorAll('.quick-btn');
	if (!discountInput || !serviceSelect || !quickButtons.length) return;

	function servicePesos() {
		const opt = serviceSelect.options[serviceSelect.selectedIndex];
		if (!opt) return 0;
		const cents = parseInt(opt.getAttribute('data-price') || '0', 10);
		return isNaN(cents) ? 0 : Math.floor(cents / 100);
	}

	quickButtons.forEach((btn) => {
		btn.addEventListener('click', () => {
			const action = btn.getAttribute('data-discount');

			if (action === 'clear') {
				discountInput.value = '0';
			} else if (action === 'free') {
				discountInput.value = String(servicePesos());
			} else {
				const n = parseInt(action, 10);
				discountInput.value = String(isNaN(n) ? 0 : n);
			}

			discountInput.dispatchEvent(new Event('input', { bubbles: true }));
			discountInput.focus();
			discountInput.select();
		});
	});
})();

// 15. App pages — animated focus ring on inputs.
// Adds a soft glow to the outline when the field is focused,
// using a CSS class so the styling lives in style.css.
(function () {
	const selectors = [
		'input:not([type="checkbox"]):not([type="radio"])',
		'select',
		'textarea',
	].join(',');

	document.querySelectorAll(selectors).forEach((el) => {
		el.addEventListener('focus', () => el.classList.add('has-focus'));
		el.addEventListener('blur',  () => el.classList.remove('has-focus'));
	});
})();