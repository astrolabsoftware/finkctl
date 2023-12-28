ARG spark_py_image

# Build stage
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o bin

# Final stage
FROM ${spark_py_image}
COPY --from=builder /app/bin /bin
CMD ["/bin"]
