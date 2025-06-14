# base python image for custom image
FROM ubuntu:latest

RUN apt-get update && apt-get install -y python3 python3-pip
RUN apt install ffmpeg libsm6 libxext6 -y

# create working directory and install pip dependencies
WORKDIR /app
COPY requirements.txt requirements.txt
RUN pip3 install -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple --break-system-packages

# copy python project files from local to /hello-py image working directory
COPY . .

CMD ["python3", "main.py"]
