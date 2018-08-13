FROM bakape/meguca
ENV PATH="${PATH}:/usr/local/go/bin"
RUN mkdir -p /hydron
WORKDIR /hydron
COPY . .
RUN make all
