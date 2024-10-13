# vulnerable-go

## Build the project

```bash
make
```

## Configure the host

Since we are working with local hosts as a mocked storage server and we want to use subdomains in the host name we need to explicitly add this to the `/etc/hosts` file.

```bash
sudo echo "127.0.0.1 3971533981712.localhost" >> /etc/hosts
```

Another option is to use te public domain name `localtest.me` for this purpose, such as `http://<something>.localtest.me`. If so, you need to change the `baseHost` variable in the `main.go` file to `localtest.me` and re-build.

## Run the mock storage server

We spin off a mocked storage server that listens on `localhost:8080` and serves the following endpoints:

- `GET /storage/{id}`: Returns the metadata of the file with the given id.
- `GET /storage/static/{id}`: Returns the content of the file with the given id.

Execute the following commands to start the server:

```bash
cd fixtures/http
npm run start
```

## Prepare an uploads directory

Create the directory `uploads` in the root of the project.

```bash
mkdir uploads/
```

We're now able to temporary store files in this directory for processing with the Go program.

## Run the Go program

Now that the storage server is running in the background we run the Go program.

## Run the project

```bash
./convert.out
```
