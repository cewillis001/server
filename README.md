Simple TFTP Server 

Instructions to Build:

If you don't already have Go installed, follow these instructions: https://golang.org/doc/install

Unzip the files and put the tftp and server folders in the directory you usually keep Go packages. 
If you went all the way through those installation instructions above, the directory structure
should look like $GOPATH/src/github.com/user/server and $GOPATH/src/github.com/user/tftp where
user is your github username.

From the command line, navigate into the server directory and enter the command "go install" to install.
Then enter $GOPATH/bin/server to start the server program.

Open another command line window, navigate into the server directory, and enter the command 
"go test" to run the included tests.

Test Descriptions:

TestOneExistsRRQ sends a single read request for an existing file

TestOneNoExistsRRQ sends a single read request for a non-existent file

TestManyExistsRRQ spawns several go routines all requesting an existing file

TestManyNoExistsRRQ spawns several go routines all requesting a non-existent file

TestManyRRQ spawns several go routines requesting files that may or may not exist

TestOneWRQ sends a single write request to the server, then a read request for that file

TestManyWRQ spawns several go routines requesting to write to the server, then read those files
