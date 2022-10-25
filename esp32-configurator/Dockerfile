FROM alpine:latest

RUN apk add --no-cache \
    python3 \
    py3-pip

ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY . .

RUN pip3 install -r /app/requirements.txt

ENTRYPOINT ["python3"]