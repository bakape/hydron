const extensions = ["jpg", "png", "gif", "webp", "pdf", "bmp", "psd", "tiff", "ogg", "webm", "mkv", "mp4", "avi", "mov", "wmv", "flv"];

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
	// Prevent defaults
	for (const e of ["dragenter", "dragexit", "dragover"]) {
		document.addEventListener(e, stopDefault);
	}
	const cont = document.getElementById("browser");

	// Set drag contents to seleceted images
	cont.addEventListener("dragstart", e => {
		let el = e.target;
		if (!el.closest || !(el = el.closest("label"))) {
			return;
		}

		const url = location.origin
			+ "/files/"
			+ el.getAttribute("data-sha1")
			+ "."
			+ extensions[parseInt(el.getAttribute("data-type"))];
		e.dataTransfer.setData("text/uri-list", url);
		console.log(url);
	});

	cont.addEventListener("mousedown", e => {
		let el = e.target;
		if (!el.closest || !(el = el.closest("label"))) {
			return;
		}
		selectImage(el);
	});

	document.addEventListener("drop", e => {
		const { files } = e.dataTransfer;
		if (!files.length || isFileInput(e.target)) {
			return;
		}
		e.stopPropagation();
		e.preventDefault();

		const browser = document.getElementById("browser");
		let done = 0;
		browser.innerHTML = "";
		for (const f of files) {
			process(f).catch(alert);
		}

		async function process(f) {
			const body = new FormData();
			body.append("file", f);
			body.append("fetch_tags", "true");
			const r = await fetch("/api/images/", { body, method: "POST" });
			if (r.status !== 200) {
				throw await r.text();
			}
			const ch = renderThumbnail(await r.json());
			browser.appendChild(ch);
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

	function selectImage(el) {
		const sel = window.getSelection();
		sel.removeAllRanges();
		const r = document.createRange();
		r.selectNodeContents(el);
		sel.addRange(r);
	}
})();

function renderThumbnail({ sha1, type, thumb: { width, height, is_png } }) {
	const cont = document.createElement("div");
	cont.innerHTML = `<label data-type="${type}" data-sha1="${sha1}">
	<input type="checkbox" name="img:${sha1}">
	<div class="background"></div>
	<img width="${width}" height="${height}" src="${thumbPath(sha1, is_png)}">
</label>`;
	return cont.firstChild;
}

function thumbPath(sha1, is_png) {
	return `/thumbs/${sha1}.${is_png ? "png" : "jpg"}`;
}

function renderProgress(val) {
	if (val === 1) {
		val = 0;
	}
	document.getElementById("progress-bar").style.width = val * 100 + "%";
}
