FROM golang:1.16.2

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/awgreene/collect-profiles

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN make build

# Run the executable
CMD [""]
