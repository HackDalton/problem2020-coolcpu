FROM golang:1.14-alpine AS build

WORKDIR /go/src/github.com/HackDalton/coolcpu
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

FROM alpine

COPY --from=build /go/bin/coolcpu ./
COPY --from=build /go/src/github.com/HackDalton/coolcpu/templates templates/
COPY --from=build /go/src/github.com/HackDalton/coolcpu/assets assets/

ENTRYPOINT ["./coolcpu"]