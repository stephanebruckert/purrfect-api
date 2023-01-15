FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN ls

RUN go build -o /purrfect-api .

EXPOSE 3000

CMD [ "/purrfect-api" ]