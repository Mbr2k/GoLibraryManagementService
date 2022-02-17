## About ##
This is a containerised microservice accepting HTTP requests with JSON bodies to interact with a library database backend.

The only manual setup required is to create a database named "library" on the mysql pod which is launched automatically.

Users' passwords are salted and hashed when stored on the database for security.

I also tried to use as few dependent packages as possible with the time available, so it is pretty lightweight.

## How to use ##
The powershell script at the top-level of this repo can be invoked on a Windows machine with "./launch.ps1" if the proper settings are enabled.

This will kick off a docker build command to generate our container's image from the Dockerfile, followed by a kubernetes declaration to run a 
mysql instance on port 3306, with default username "root" and password "dev".

The docker run command will then be invoked which will publish the exposed port 8080 of the running container, and target the mysql database 
with port "3306", 
now running on "db.localhost:3306" 
- this db.localhost variable can be changed to whatever machine's external IP your mysql database is running on.

The shell will then hook in to the go/src directory from where a "go build" command can be invoked to build the module in to the container locally or otherwise run the application.

A default superuser/root password environment variable is also set as libraryRootPassword="Alexandria".

The Postman JSON collection at the top level directory of this repo also details the different endpoints with example calls.
