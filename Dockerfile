FROM bakape/meguca
ENV GOPATH=/go
ENV PATH="${PATH}:/usr/local/go/bin:${GOPATH}/bin"
RUN mkdir -p /go/src/github.com/bakape/hydron
WORKDIR /go/src/github.com/bakape/hydron
COPY . .
RUN npm install
RUN go get -v ./...
RUN make all
