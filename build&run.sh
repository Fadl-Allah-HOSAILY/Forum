# Build Docker image
docker build -t forum-app .

# Run container
docker run -d -p 8081:8081 --name forum-container forum-app
