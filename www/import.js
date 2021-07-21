const browser = document.getElementById("browser");

// Path import
(() => {
	const form = document.getElementById("import");
	
	document.getElementById("submit").addEventListener("click", async () => {
        if (!confirm("Generic confirmation message.")) {
            return;
		}
		input = form.querySelector("#path").value;
		if (input.length === 0) {
			alert("Must enter an import path.");
			return;
		}

        const body = "path=" + input +
            "&del=" + form.querySelector("#delete").checked +
            "&fetchTags=" + form.querySelector("#fetch-tags").checked +
			"&storeName=" + form.querySelector("#store-name").checked +
			"&tagStr=" + form.querySelector("#input-tags").value;

		let r = await fetch("/api/import", { body, method: "POST",
			headers: { "Content-Type": "application/x-www-form-urlencoded" } } );
		const reader = r.body.getReader();
		const decoder = new TextDecoder("utf-8");

		// Recursively read from stream and process chunks,
		// until "done" message
		await read();
		async function read() {
			let chunk = await reader.read();
			if (chunk.done) {
				return;
			}
			// Sometimes server sends multiple chunks before client
			// finishes processing, so have to split them
			s = decoder.decode(chunk.value).split("-");
			for (let i = 0; i < s.length - 1; i++) {
				let obj = JSON.parse(s[i]);
				await addThumb(obj.SHA1);
				renderProgress(obj.Current / obj.Total);
			}
			await read();
		}
    }, { passive: true });
})();

// Drag and drop import
(() => {
    // Prevent defaults
	for (const e of ["dragenter", "dragexit", "dragover"]) {
		document.addEventListener(e, stopDefault);
	}
	
	async function process(f) {
		const body = new FormData();
		body.append("file", f);
		body.append("fetch_tags", "true");
		body.append("store_name", "true");
		let r = await fetch("/api/images/", { body, method: "POST" });
		if (r.status !== 200) {
			throw await r.text();
		}
		
		await addThumb((await r.json()).sha1);
	}
	
	// Properly reload page when going back through history, after drag&drop
	// redirect from main page
	window.onpopstate = function() {
		window.location.assign(window.location.href);
	}
	
	// Check if there is any drag&drop data saved in history to import
	window.onload = async function() {
		if (history.state !== null) {
			let done = 0;
			for (const f of history.state) {
				await process(f).catch(alert);
				renderProgress(++done / history.state.length);
			}
		}
	}
    
	document.addEventListener("drop", async e => {
		const { files } = e.dataTransfer;
		if (!files.length || isFileInput(e.target)) {
			return;
		}
		preventDefault(e);
		let done = 0;
		for (const f of files) {
			await process(f).catch(alert);
			renderProgress(++done / files.length);
		}
    });
	
	function preventDefault(e) {
		e.stopPropagation();
		e.preventDefault();
	}

    function stopDefault(e) {
		if (!isFileInput(e.target)) {
			preventDefault(e);
		}
	}

	function isFileInput(el) {
		return el.tagName === "INPUT" && el.getAttribute("type") === "file";
	}
})();

function renderProgress(val) {
	if (val === 1) {
		val = 0;
	}
	document.getElementById("progress-bar").style.width = val * 100 + "%";
}

async function addThumb(hash) {
	let r = await fetch("/ajax/thumbnail/" + hash);
	if (r.status !== 200) {
		throw await r.text();
	}
	const cont = document.createElement("div");
	cont.innerHTML = await r.text();
	browser.appendChild(cont.firstChild);
}