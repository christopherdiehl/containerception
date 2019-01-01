# STEP 1 build executable binary
FROM golang:alpine 

COPY containerception /usr/local/bin

CMD ["/bin/ash"]