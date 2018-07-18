const browser = document.getElementById("browser");
const imageView = document.getElementById("image-view");

// Search suggestions
(() => {
	const search = document.getElementById("search");
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
		e.stopPropagation();
		e.preventDefault();

		let done = 0;
		browser.innerHTML = "";
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
		}
	});

	function stopDefault(e) {
		if (!isFileInput(e.target)) {
			e.stopPropagation();
			e.preventDefault();
		}
	}

	function isFileInput(el) {
		return el.tagName === "INPUT" && el.getAttribute("type") === "file";
	}
})();

window.onhashchange = e =>
	loadHash(e.newURL);
loadHash(location.toString(), true); // On page load

function loadHash(url, firstLoad) {
	const hash = new URL(url).hash;
	if (hash.startsWith("#img:")) {
		viewImage(hash.slice(5));
	} else {
		imageView.innerHTML = "";
	}
}

browser.addEventListener("click", e => {
	if (e.target.closest && e.target.tagName !== "INPUT") {
		viewImage(e.target.closest("figure").getAttribute("data-sha1"));
	}
}, { passive: true });

document.addEventListener("keydown", e => {
	if (e.key === "Escape" && imageView.innerHTML !== "") {
		history.back();
	}
}, { passive: true });

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
}
