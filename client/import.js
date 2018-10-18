// Path import
(() => {
    const form = document.getElementById("import");
    const sub = document.getElementById("submit");

    sub.addEventListener("click", async e => {
        if (!confirm("Generic confirmation message")){
            return;
        }

        const body = "path=" + form.querySelector("#path").value +
            "&del=" + form.querySelector("#delete").checked +
            "&fetchTags=" + form.querySelector("#fetch-tags").checked +
            "&tagStr=" + form.querySelector("#input-tags").value;
        try {
            const r = await fetch("/api/import", { body, method: "POST", 
            headers: { "Content-Type": "application/x-www-form-urlencoded"}
            });
        } catch(err) {
            alert(err);
        }
    }, { passive: true });
})();