docker build . -t go-library-service;
kubectl apply -f mysql.yaml;
docker run --rm -it  -v "$($PWD):/go:z" -v "$($env:USERPROFILE)\.ssh:/tmp/.ssh:z" -h libraryserver --name libraryservice -e DB_USER="root" -e DB_PW="dev" -e DB_URL="db.localhost:3306" -e DB_NAME="library" -e libraryRootPassword="Alexandria" -p 8080:8080  -t go-library-service
