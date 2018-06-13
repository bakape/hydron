function get(url, cb) {
    var xhr = new XMLHttpRequest()
    xhr.onreadystatechange = function () {
        if (xhr.readyState === xhr.DONE) {
            if (xhr.status === 200) {
                cb(JSON.parse(xhr.responseText), null)
            } else {
                cb(null,
                   xhr.status + ": " + (xhr.responseText || "server unavailable"))
            }
        }
    }
    xhr.open("GET", "http://localhost:8010/api" + url, true)
    xhr.send()
}
