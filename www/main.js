const extensions = ["jpg", "png", "gif", "webp", "pdf", "bmp", "psd", "tiff", "ogg", "webm", "mkv", "mp4", "avi", "mov", "wmv", "flv"];

// Search suggestions
(() => {
	const search = document.getElementById("search");
	const sugg = document.getElementById("search-suggestions");
	if (!search) {
		return;
	}

	search.addEventListener("input", e => {
		const text = search.value;
		if (!text.length || text[text.length - 1] == " ") {
			sugg.innerHTML = "";
			return;
		}

		const i = text.lastIndexOf(" ");
		fetch(`/api/complete_tag/${i === -1 ? text : text.slice(i + 1)}`)
			.then(r => r.json())
			.then(tags => {
				let text = search.value;
				let i = text.lastIndexOf(" ");
				if (i == -1) {
					i = 0;
				}
				text = text.slice(0, i);
				if (i) {
					text += " ";
				}
				let s = "";
				for (const tag of tags) {
					s += `<option value="${text} ${tag}">`;
				}
				sugg.innerHTML = s;
			})
			.catch(alert);
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

	function stopDefault(e) {
		const el = e.target;
		if (!(el.tagName === "INPUT" && el.getAttribute("type") === "file")) {
			e.stopPropagation();
			e.preventDefault();
		}
	}

	function selectImage(el) {
		const sel = window.getSelection();
		sel.removeAllRanges();
		const r = document.createRange();
		r.selectNodeContents(el);
		sel.addRange(r);
	}
})();
