// Path import
(() => {
    const form = document.getElementById("import");
    const sub = document.getElementById("submit");

    sub.addEventListener("click", async e => {
        if (!confirm("Generic confirmation message")){
            return;
		}
		input = form.querySelector("#path").value;
		if (input.length === 0){
			return;
		}

        const body = "path=" + input +
            "&del=" + form.querySelector("#delete").checked +
            "&fetchTags=" + form.querySelector("#fetch-tags").checked +
            "&tagStr=" + form.querySelector("#input-tags").value;
        try {
            const r = await fetch("/api/import", { body, method: "POST", 
            headers: { "Content-Type": "application/x-www-form-urlencoded" }
            });
        } catch(err) {
            alert(err);
        }
    }, { passive: true });
})();

// Drag and drop import
(() => {
    const browser = document.getElementById("browser");
    // Prevent defaults
	for (const e of ["dragenter", "dragexit", "dragover"]) {
		document.addEventListener(e, stopDefault);
    }
    
	document.addEventListener("drop", e => {
		const { files } = e.dataTransfer;
		if (!files.length || isFileInput(e.target)) {
			return;
		}
		preventDefault(e);
		let done = 0;
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
			preventDefault(e);
		}
	}

	function isFileInput(el) {
		return el.tagName === "INPUT" && el.getAttribute("type") === "file";
	}
})();

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
