FROM golang:1.19
LABEL vesion="1.0"
LABEL maintaner="Aliaksei Vidaseu"
LABEL description="Social Network Backend"
LABEL port="8080"
WORKDIR /server
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build -o /docker-social-network-backend
CMD [ "/docker-social-network-backend" ]