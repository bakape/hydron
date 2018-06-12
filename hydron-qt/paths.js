
function thumb(sha1, isPNG) {
    return "http://localhost:8010/thumbs/" + sha1 + "." + (isPNG ? "png" : "jpg")
}

function source(sha1, fileType) {
    var s = "http://localhost:8010/files/" + sha1 + "."
    switch (fileType) {
    case 0:
        s += "jpg"
        break
    case 1:
        s += "png"
        break
    case 2:
        s += "gif"
        break
    case 3:
        s += "webp"
        break
    case 4:
        s += "pdf"
        break
    case 5:
        s += "bmp"
        break
    case 6:
        s += "psd"
        break
    case 7:
        s += "tiff"
        break
    case 8:
        s += "ogg"
        break
    case 9:
        s += "webm"
        break
    case 10:
        s += "mkv"
        break
    case 11:
        s += "mp4"
        break
    case 12:
        s += "avi"
        break
    case 13:
        s += "mov"
        break
    case 14:
        s += "wmw"
        break
    case 15:
        s += "flv"
        break
    }
    return s
}
