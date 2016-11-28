# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/rwapps/headlines_most_read

# Build the app and dependencies inside the container.
RUN go get "golang.org/x/net/context"
RUN go get "golang.org/x/oauth2/jwt"
RUN go get "google.golang.org/api/analyticsreporting/v4"
RUN go install github.com/rwapps/headlines_most_read

ENTRYPOINT /go/bin/headlines_most_read
