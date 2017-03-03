This file documents guidelines and specifications for developing client
applications for Hydron.

##Binary path

Windows clients should likely be packaged with a statically compiled Hydron
binary, downloaded from the release page and resolve the path to it. Linux
clients can simply assume the `hydron` command is inside the system `$PATH`.

##Startup

The client can either launch it's own instance of Hydron with `hydron serve` or
connect to a running remote instance over the network. When launching as a child
process with an already running instance using the database, the child will lock
in wait for the database to unlock. This lock will not reliably timeout on
Windows, so the client is free to display an error message after an arbitrary
amount of time.

##API
The server will listen for requests through an HTTP API described bellow.
Default port is 8010.

| URL | Method | Parameters | Response Type | Description |
|--|---|---|---|---|
| /get/:IDs[?minimal=true] | GET | IDs: a comma-separated list of one or more hex-encoded file SHA1 hashes. If the `minimal=true` query string is passed, the returned records will not include the MD5 or tags fields. This can significantly improve response times. | [Record](#record)[] | Returns records with the target IDs. Unmatched records are ignored.  If :IDs is empty, returns all records in the database. |
| /search/:tags[?minimal=true] | GET | tags: a comma-separated list of [tags](#tag). If the `minimal=true` query string is passed, the returned records will not include the MD5 or tags fields. This can significantly improve response times. | [Record](#record)[] | Returns records that match all the provided tags.  If  :tags is empty, returns all records in the database. |
| /complete_tag/:prefix | GET | prefix: string that you would like to autocomplete | string[] | Returns up to the first ten tags that start with :prefix. |
| /files/:file | GET | :file hex-encoded file SHA1 hash followed by dot and extension | file | Returns the source file specified by :file.**\*** |
| /thumbs/:file | GET | :file hex-encoded file SHA1 hash followed by dot and extension | file | Returns the thumbnail image specified by :file.**\*** |
| /import[?fetch_tags=true] | POST | A form with the file to import in the "file" field. If the `fetch_tags=true` query string is passed, Hydron will attempt to fetch tags from gelbooru.com right after importing.  | - | Uploads a file to be imported into the database. |
| /fetch_tags | POST | - | - | Instructs Hydron to attempt to fetch tags from gelbooru.com for all eligible files, that have not had their tags fetched yet. |
| /remove/:IDs | POST | IDs: a comma-separated list of one or more hex-encoded file SHA1 hashes. | - | Remove the files specified by :IDs. |
| /add_tags/:ID/:tags | POST | ID: hex-encoded SHA1 hash of the file you want to add tags to tags: a comma-separated list of [tags](#tag) | - | Add the specified tags to the specified file |
| /remove_tags/:ID/:tags | POST | ID:hex-encoded SHA1 hash of the file you want to remove tags from tags: a comma-separated list of [tags](#tag) | - | Remove the specified tags from the specified file |

###Files

**\*** If the client launched Hydron itself as a child, there is no need to use
HTTP for retrieving file assets. All source images are located on the disk
by the path
`$HOME/.hydron/images/<first two letters of hash>/<hash>.<extension>` on POSIX
systems and
`%APPDATA%\hydron\images\<first two letters of hash>\<hash>.<extension>` on
Windows.

Similarly thumbnails can be found by
`$HOME/.hydron/thumbs/<first two letters of hash>/<hash>.<extension>` and
`%APPDATA%\hydron\images\<first two letters of hash>\<hash>.<extension>`
respectively.

The extension of thumbnails will be `jpg`, unless the `thumbIsPNG` field of
[Record](#record) is true, in which case the extension will be `png`.

Note, that you should never modify these files directly. Doing so may corrupt
the database.

##Types

###Record
Contains information about a single file stored in the database.

| Field | Type | Description |
|---|:---:|---|
| SHA1 | string | Hex-encoded SHA1 hash of the source file. Used as the image ID throughout Hydron. |
| type | string | File type of the source image (jpg, png, mkv, ...). |
| thumbIsPNG | bool | Specifies, if the thumbnail of the file is a PNG image. If false, the thumbnail is a JPEG image.  |
| importTime | uint | Unix timestamp of the time the file was imported into Hydron. |
| size | uint | Source file size in bytes. |
| width | uint | Source file width in pixels. |
| height | uint | Source file height in pixels. |
| length | uint | Length of source file in seconds . Only relevant for video and audio files. |
| MD5 | string | Hex-encoded MD5 hash of the source file. |
| tags | string[] | Array of the file's [tags](#tag). |

###tag
A tag describes a property the file has or a category it belongs to.

Tags contain no spaces and use underscores, where otherwise spaces would be
used. For performance reasons Hydron does not have tag namespaces and all tags
are in the same namespace. Any input namespaced tags (`namespace:tag`), except
for the exceptions below, will be stripped of their namespace.

####Rating tags
Tags that define a safety rating of an image.

These are one of `rating:safe`, `rating:questionable` or `rating:explicit`.
The respective shorthands `rating:s`, `rating:q` and `rating:e` are also
accepted.

####System tags
System tags allow searching by file metadata.

These are prefixed with `system:` and followed by
`size`, `width`, `height`, `length`, `tag_count`,
followed by one of these comparison operators:
`>`, `<`, `=`
and a positive number.
Examples:
`system:width>1920` `system:height>1080` `system:tag_count=0`
`system:size<10485760`
