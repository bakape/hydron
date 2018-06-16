// Search suggestions
(function () {
	var search = document.getElementById("search");
	var sugg = document.getElementById("search-suggestions");
	if (!search) {
		return;
	}

	search.addEventListener("input", function (e) {
		var text = search.value;
		if (!text.length || text[text.length - 1] == " ") {
			sugg.innerHTML = "";
			return
		}

		var i = text.lastIndexOf(" ");
		var last = i === -1 ? text : text.slice(i + 1);
		HTTPGet("/api/complete_tag/" + last, function (s) {
			var tags = JSON.parse(s);
			s = "";
			var text = search.value;
			var i = text.lastIndexOf(" ");
			if (i == -1) {
				i = 0;
			}
			text = text.slice(0, i);
			if (i) {
				text += " ";
			}
			for (var i = 0; i < tags.length; i++) {
				s += "<option value=\"" + text + tags[i] + "\">";
			}
			sugg.innerHTML = s;
		});
	}, { passive: true });
})();

function HTTPSend(url, method, body, cb) {
	var xhr = new XMLHttpRequest();
	xhr.open(method, url, true);
	xhr.onload = function () {
		if (this.status !== 200) {
			alert(this.responseText);
		} else {
			cb(this.responseText);
		}
	};
	xhr.send(body);
}

function HTTPGet(url, cb) {
	HTTPSend(url, "GET", undefined, cb);
}
