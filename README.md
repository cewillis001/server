Simple TFTP Server 

Instructions to Build:

If you don't already have Go installed, follow these instructions: https://golang.org/doc/install

If you don't already have copies of the server and tftp folders, you can download archives of them from https://github.com/cewillis001/server and https://github.com/cewillis001/tftp, respectively.

Unzip the files and put the tftp and server folders in the directory you usually keep Go packages. 
If you went all the way through those installation instructions above, the directory structure
should look like $GOPATH/src/github.com/user/server and $GOPATH/src/github.com/user/tftp where
user is your github username.

From the command line, navigate into the server directory and enter the command "go install" to install. Then enter $GOPATH/bin/server to start the server program. Setting the verbose flag ($GOPATH/bin/server -verbose) will output the server's activities in detail. Open another command line window, navigate into the user/server directory, and enter the command "go test" to run the included tests. There is a variable in server_test.go that can be changed edited to test concurrency.

Notes on Implementation:

This is a TFTP server ( http://tools.ietf.org/html/rfc1350 ), implimenting only octet mode, which means it stores data as an array of bytes. Files in the process of being written are not visible to clients requesting reads until after the entire file has been sucessfully transmitted. Files are stored in a map in memory, so they don't persist after the server is closed. The max effective concurrency seems to be around 1000 go routineson my machine around which timeouts become very frequent. I'm running Ubuntu 14.04 x64 with 512 Mb RAM. This performance might be increased by implimenting channels with larger buffers for the bottleneck functions WriteFile in server.go, as well as by implimenting an adaptive timeout function that expands to account for a particularly busy network.

Note: an earlier version had the max effective concurrency at around 200; this was being caused by runtime panics on trying to send on a closed channel. I folded the implimentation of reading and writing to the map of open channels serving read and write requests into its own goroutine, which fixed this problem.
