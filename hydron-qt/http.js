function get(url, cb) {
    send(url, "GET", undefined, cb)
}

function post(url, body, cb) {
    send(url, "POST", body, cb)
}

function send(url, method, body, cb) {
    var xhr = new XMLHttpRequest()
    xhr.onreadystatechange = function () {
        if (xhr.readyState === xhr.DONE) {
            if (xhr.status === 200) {
                cb(xhr.responseText ? JSON.parse(xhr.responseText) : "", null)
            } else {
                cb(null, xhr.status + ": " + (xhr.responseText
                                              || "server unavailable"))
            }
        }
    }
    xhr.open(method, "http://localhost:8010/api" + url, true)
    xhr.send(body)
}
