(() => {
  let socket = new WebSocket("ws://127.0.0.1:9898/ws");
  console.log("Attempting Connection...");

  socket.onopen = () => {
    console.log("Successfully Connected");
    socket.send("Hi!! This message is From the Client!")
  };

  socket.onmessage = event => {
    console.log(`Recieved Message: ${event.data}`)
    if (event.data == 1) {
      socket.send("Rebuild & Reload")
      location.reload();
    }
  }

  socket.onclose = event => {
    console.log("Socket Closed Connection: ", event);
    socket.send("Client Closed!")
  };

  socket.onerror = error => {
    console.log("Socket Error: ", error);
  };
})();
