FROM golang:1.19
LABEL vesion="1.0"
LABEL maintaner="Aliaksei Vidaseu"
LABEL description="Social Media Backend"
LABEL port="8080"
WORKDIR /server
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build -o /docker-social-media-backend
CMD [ "/docker-social-media-backend" ]