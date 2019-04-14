// Path import
(() => {
    const form = document.getElementById("import");
    const sub = document.getElementById("submit");

    sub.addEventListener("click", () => {
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
			"&tagStr=" + form.querySelector("#input-tags").value;

		fetch("/api/import", { body, method: "POST",
			headers: { "Content-Type": "application/x-www-form-urlencoded" } } )
		.then(response => {
			return new Response(
				new ReadableStream({
					start(controller) {
						const reader = response.body.getReader();
						const decoder = new TextDecoder("utf-8");

						// Recursively read from stream and process chunks,
						// until "done" message
						read();
						function read() {
							reader.read().then(({done, value}) => {
								if(done) {
									controller.close();
									return;
								}
								// Sometimes server sends multiple chunks before client
								// finishes processing, so have to split them
								s = decoder.decode(value).split("-");
								for(i = 0; i < s.length - 1; i++) {
									var obj = JSON.parse(s[i]);

									fetch(`/ajax/thumbnail/${obj.SHA1}`)
									.then( r => {
										if (r.status !== 200) {
											r.text().then( r => {
												throw r;
											})
										}
										r.text().then( r => {
											const cont = document.createElement("div");
											cont.innerHTML = r;
											browser.appendChild(cont.firstChild);

											renderProgress(obj.Current / obj.Total)
										})
									})
								}
								read();
							})
						}
					}
				})
			);
		})
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
