const browser = document.getElementById("browser");
const imageView = document.getElementById("image-view");
const search = document.getElementById("search");
const figureWidth = 200 + 4; // With margin

// Search bar and suggestions
(() => {
	const sugg = document.getElementById("search-suggestions");
	if (!search) {
		return;
	}

	// Autosubmit on order change
	const form = document.querySelector("#top-banner form")
	form.addEventListener("change", e => {
		switch (e.target.getAttribute("name")) {
			case "order":
			case "reverse":
				form.submit();
		}
	}, { passive: true });


	search.addEventListener("input", async () => {
		let text = search.value;
		if (!text.length || text[text.length - 1] == " ") {
			sugg.innerHTML = "";
			return;
		}

		try {
			let i = text.lastIndexOf(" ");
			const r = await fetch("/api/complete_tag/"
				+ (i === -1 ?encodeURIComponent(text) :
				encodeURIComponent(text.slice(i + 1))));
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
	// Prevent defaults
	for (const e of ["dragenter", "dragexit", "dragover"]) {
		document.addEventListener(e, stopDefault);
	}

	// Set drag contents to selected images
	browser.addEventListener("dragstart", e => {
		let el = e.target;
		if (!el.closest || !(el = el.closest("figure"))) {
			return;
		}
		e.dataTransfer.setData("text/uri-list",
			location.origin + el.getAttribute("data-href"));
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

	// Clear state to avoid reimporting, and properly reload page when
	// going forward through history, after drag&drop redirect
	window.onpopstate = function() {
		history.replaceState(null, "");
		window.location.assign(window.location.href);
	}

	document.addEventListener("drop", e => {
		const { files } = e.dataTransfer;
		if (!files.length || isFileInput(e.target)) {
			return;
		}
		preventDefault(e);

		history.pushState(files, "", "/import");
		window.location.assign("/import");
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

// Options form
(() => {
	const optsform = document.querySelector("#opts-bar");
	const optssub = document.querySelector("#opts-submit");

	optssub.addEventListener("click", async () => {
		if (!confirm("Generic confirmation message")){
			return;
		}

		let checked = [];
		for (let el of [...browser.children]) {
			el = el.querySelector("input[type=checkbox]");
			if (el.checked) {
				let i = el.name.lastIndexOf(":");
				checked.push(el.name.slice(i+1));	// Extract+store image IDs
			}
		}

		let opt = optsform.querySelector("#opts-select").value;
		let input = optsform.querySelector("#opts-input").value;
		params = createParams(opt, input);
		if (params.length === 0){
			return;
		}

		let n = checked.length;
		let i = 0;
		while (i < n) {
			try {
				let path = "/api/images/" + checked[i] + params[0];
				await fetch( path, { method: params[1],
					 body: params[2],
					 headers: {
						 "Content-Type": "application/x-www-form-urlencoded"
						} });
			} catch (err) {
				alert(err);
			}
			renderProgress(++i/n);
		}
	}, { passive: true });
})();

document.addEventListener("keydown", e => {
	if (e.getModifierState("Alt")) {
		return;
	}
	if (e.path.includes(search)) {
		if (e.getModifierState("Control") && e.key === "b") {
			// Remove focus from search bar so other keybinds work
			search.blur();
			preventDefault(e);
		}
		return;
	}
	let matched = true;
	let h;
	if (e.getModifierState("Control")) {
		switch (e.key) {
			case "l":
				search.focus();
				break;
			case "a":
				for (let el of [...browser.children]) {
					el = el.querySelector("input[type=checkbox]");
					el.checked = !el.checked;
				}
				break;
			default:
				matched = false;
		}
	} else {
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
				getHighlighted().querySelector("a").click();
				break;
			case "PageDown":
				for (let c of document.getElementById("page-links").children) {
					if (c.textContent === "<") {
						c.click();
					}
				}
				break;
			case "PageUp":
				for (let c of document.getElementById("page-links").children) {
					if (c.textContent === ">") {
						c.click();
					}
				}
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
	}
	if (matched) {
		preventDefault(e);
	}
});

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

function createParams(option, input) {
	suffix = ["/tags/fetch", "/tags/", "/tags/", "/name", "/"];
	methods = ["PATCH", "PATCH", "POST", "POST", "DELETE"];
	bodys = ["", "tags=", "tags=", "name=", ""];

	if (option > suffix.length) {
		console.log("Invalid selection");
		return [];
	}

	method = methods[option];
	body = bodys[option];
	if (body.length != 0){
		body += input;
	}
	
	return [ suffix[option], method, body ];
}
