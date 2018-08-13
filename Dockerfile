FROM bakape/meguca
ENV PATH="${PATH}:/usr/local/go/bin"
RUN mkdir -p /go/src/github.com/bakape/hydron
WORKDIR /go/src/github.com/bakape/hydron
ENV GOPATH=/go
COPY . .
RUN npm install
RUN go get .
RUN make all
