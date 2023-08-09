FROM python:3.10-slim

WORKDIR /app

RUN export DEBIAN_FRONTEND=noninteractive \
    && apt-get -qq update \
    && apt-get -qq install --no-install-recommends \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt requirements.txt
RUN pip3 install -r requirements.txt
RUN pip3 install python-multipart

COPY . .

EXPOSE 8000
CMD ["python3", "main.py"]