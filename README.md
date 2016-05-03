## inMemoryServer.go ##
A simple http file server written in go that loads all files from the `www/` directory into memory, and serves them from there to clients.

### Why!? ###
* Sometimes all you want to do is serve a few large files to your users
* Disk access can be expensive (in comparison to memory) when you're only serving a few files, but those a lot of times

### How To Use ###
1. Create a `www/` directory and put files there as needed
2. Run the server: `go run inMemoryServer.go`
3. Content will be accessible at `localhost:9090` (port can be changed in the .go file)