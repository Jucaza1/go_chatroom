### go_chatroom

**A server providing 2 endpoints:**
-A web page acting as a client to actively connect to through websocket
-An endpoint to recieve request to establish pasively websocket connection

**Besides the http server there are 3 kind of goroutines:**
-1 for the Hub to keep track of the clients, and broadcast to them
-2 for each client, one for reading and one for writing
