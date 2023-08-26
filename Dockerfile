FROM golang:1.19.0-alpine3.16 AS build

# The /app directory should act as the main application directory
WORKDIR /app

# Copy the app package and package-lock.json file
COPY db ./
COPY go ./
COPY go.mod ./
COPY go.sum ./
COPY main.go ./
COPY main_test.go ./

EXPOSE 8080

# Start the app using serve command
CMD [ "serve", "-s", "build" ]