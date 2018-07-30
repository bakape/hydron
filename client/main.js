const browser = document.getElementById("browser");
const imageView = document.getElementById("image-view");
const search = document.getElementById("search");
const figureWidth = 200 + 4; // With marging

// Search suggestions
(() => {
	const sugg = document.getElementById("search-suggestions");
	if (!search) {
		return;
	}

	search.addEventListener("input", async e => {
		let text = search.value;
		if (!text.length || text[text.length - 1] == " ") {
			sugg.innerHTML = "";
			return;
		}

		try {
			let i = text.lastIndexOf(" ");
			const r = await fetch("/api/complete_tag/"
				+ (i === -1 ? text : text.slice(i + 1)));
			if (r.status !== 200) {
				throw await r.text();
			}

			const tags = await r.json();
			if (i === -1) {
				i = 0;
			}
			text = text.slice(0, i);
			if (text.length) {
				text += " ";
			}
			let s = "";
			for (const tag of tags) {
				s += `<option value="${text}${tag}">`;
			}
			sugg.innerHTML = s;
		} catch (err) {
			alert(err);
		}
	}, { passive: true });
})();

// Drag and drop
(() => {
	const extensions = ["jpg", "png", "gif", "webp", "pdf", "bmp", "psd",
		"tiff", "ogg", "webm", "mkv", "mp4", "avi", "mov", "wmv", "flv"];

	// Prevent defaults
	for (const e of ["dragenter", "dragexit", "dragover"]) {
		document.addEventListener(e, stopDefault);
	}

	// Set drag contents to seleceted images
	browser.addEventListener("dragstart", e => {
		let el = e.target;
		if (!el.closest || !(el = el.closest("figure"))) {
			return;
		}
		e.dataTransfer.setData("text/uri-list", location.origin
			+ "/files/"
			+ el.getAttribute("data-sha1")
			+ "."
			+ extensions[parseInt(el.getAttribute("data-type"))]);
	});

	browser.addEventListener("mousedown", e => {
		let el = e.target;
		if (!el.closest || !(el = el.closest("figure"))) {
			return;
		}

		// Select image
		const sel = window.getSelection();
		sel.removeAllRanges();
		const r = document.createRange();
		r.selectNodeContents(el);
		sel.addRange(r);
	});

	document.addEventListener("drop", e => {
		const { files } = e.dataTransfer;
		if (!files.length || isFileInput(e.target)) {
			return;
		}
		preventDefault(e);

		let done = 0;
		browser.innerHTML = search.value = "";
		for (const f of files) {
			process(f).catch(alert);
		}

		async function process(f) {
			const body = new FormData();
			body.append("file", f);
			body.append("fetch_tags", "true");
			let r = await fetch("/api/images/", { body, method: "POST" });
			if (r.status !== 200) {
				throw await r.text();
			}

			r = await fetch(`/ajax/thumbnail/${(await r.json()).sha1}`)
			if (r.status !== 200) {
				throw await r.text();
			}
			const cont = document.createElement("div");
			cont.innerHTML = await r.text();
			browser.appendChild(cont.firstChild);
			renderProgress(++done / files.length);
			if (done === 1) {
				setHighlight(browser.querySelector("figure"));
			}
		}
	});

	function stopDefault(e) {
		if (!isFileInput(e.target)) {
			preventDefault(e);
		}
	}

	function isFileInput(el) {
		return el.tagName === "INPUT" && el.getAttribute("type") === "file";
	}
})();

window.onhashchange = e =>
	loadHash(e.newURL);
loadHash(location.toString(), true); // On page load

browser.addEventListener("click", e => {
	if (!e.target.closest || e.target.tagName === "INPUT") {
		return;
	}
	viewImage(e.target.closest("figure").getAttribute("data-sha1"));
	setHighlight(e.target);
}, { passive: true });

browser.addEventListener("keydown", e => {
	let matched = true;
	let h;
	switch (e.key) {
		case "ArrowDown":
			moveHighlight(0, +1);
			break;
		case "ArrowUp":
			moveHighlight(0, -1);
			break;
		case "ArrowRight":
			moveHighlight(+1, 0);
			break;
		case "ArrowLeft":
			moveHighlight(-1, 0);
			break;
		case " ": // SpaceBar
			h = getHighlighted();
			if (h) {
				const chb = h.querySelector("input[type=checkbox]")
				chb.checked = !chb.checked;
			}
			break;
		case "Enter":
			h = getHighlighted();
			if (h) {
				viewImage(h.getAttribute("data-sha1"));
			}
			break;
		case "PageDown":
			moveHighlight(0, +browserWidth());
			break;
		case "PageUp":
			moveHighlight(0, -browserWidth());
			break;
		case "Home":
			setHighlight(browser.querySelector("figure"));
			break;
		case "End":
			setHighlight(browser.querySelector("figure:last-child"));
			break;
		default:
			matched = false;
	}
	if (matched) {
		preventDefault(e);
	}
});

imageView.addEventListener("keydown", e => {
	if (e.key === "Escape" && imageView.innerHTML !== "") {
		history.back();
		browser.focus();
	}
}, { passive: true });

function loadHash(url, firstLoad) {
	const hash = new URL(url).hash;
	if (hash.startsWith("#img:")) {
		viewImage(hash.slice(5));
	} else {
		imageView.innerHTML = "";
	}
}

// Return function, that prevents default behavior, when fn() returns true
function maybePreventDefault(fn) {
	return e => {
		if (fn()) {
			preventDefault(e);
		}
	};
}

function preventDefault(e) {
	e.stopPropagation();
	e.preventDefault();
}

function renderProgress(val) {
	if (val === 1) {
		val = 0;
	}
	document.getElementById("progress-bar").style.width = val * 100 + "%";
}

async function viewImage(sha1) {
	const r = await fetch(`/ajax/image-view/${sha1}${location.search}`);
	if (r.status !== 200) {
		alert(await r.text());
	}
	location.hash = `#img:${sha1}`;
	imageView.innerHTML = await r.text();
	imageView.focus();
}

function browserWidth() {
	return Math.floor(browser.offsetWidth / figureWidth);
}

function browserHeight() {
	return Math.floor(browser.offsetHeight / figureWidth);
}

// Returns browser grid as 2D array and the position of the highlighted figure
function browserGrid() {
	const colums = browserWidth();
	const grid = [];
	let c = 0;
	let highlighted;
	for (const f of browser.querySelectorAll("figure")) {
		if (!grid[c]) {
			grid[c] = [];
		}
		if (f.classList.contains("highlight")) {
			highlighted = { x: grid[c].length, y: c };
		}
		grid[c].push(f);
		if (grid[c].length === colums) {
			c++;
		}
	}
	return { grid, highlighted };
}

function moveHighlight(xMove, yMove) {
	let { grid, highlighted: { x, y } } = browserGrid();
	x += xMove;
	y += yMove;

	// Wrap around rows
	const bw = browserWidth();
	if (x < 0) {
		y--;
		x += bw;
	} else if (x >= bw) {
		y++;
		x -= bw;
	}

	if (!grid[y]) {
		return;
	}
	const h = grid[y][x];
	if (!h) {
		return;
	}
	setHighlight(h);
}

function getHighlighted() {
	return browser.querySelector("figure.highlight");
}

function setHighlight(target) {
	if (!target || !target.closest || !(target = target.closest("figure"))) {
		return;
	}
	const h = getHighlighted();
	if (h) {
		h.classList.remove("highlight");
	}
	target.classList.add("highlight");
	target.scrollIntoView({
		behavior: "smooth",
		block: "center",
	});
}
