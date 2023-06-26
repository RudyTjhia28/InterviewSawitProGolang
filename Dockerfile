FROM golang:1.20

WORKDIR /cmd

COPY . .

RUN go mod download

RUN go build -o interviewsawitprogolang ./cmd

EXPOSE 8080

CMD [ "./interviewsawitprogolang" ]