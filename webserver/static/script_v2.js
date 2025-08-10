let eventSource = null;
const messageDiv = document.getElementById('message');
const pingDiv = document.getElementById('serverPing');
const dcButton = document.getElementById('dcButton');

async function authenticate(username, key) {
  try {
    // Create JSON payload
    const payload = JSON.stringify({
      username: username,
      key: key
    });
    
    //send auth
    const response = await fetch('/auth', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: payload
    });
    
    if (!response.ok) {
      console.error('Authentication failed');
      return false;
    }
  
    //return auth
    const data = await response.json();
    return data.status === 'success';
  } catch (error) {
    console.error('Authentication error:', error);
    return false;
  }
}

function connectSSE() {


  // Close any existing connection
  if (eventSource) {
    eventSource.close();
  }

  var lastDate = Date.now();
  eventSource = new EventSource('/sse');
  
  eventSource.onopen = function() {
    //ping logic
    const now = Date.now();
    var ping = now - lastDate
    ping = ping / 2
    console.log(`Ping: ${ping}ms`);
    pingDiv.textContent = `Ping: ${ping}ms`;
    pingDiv.style.color = '#064a72';
    //----

    messageDiv.innerText = 'Successfully authenticated!\nDO NOT Close this window or the session will end';
    messageDiv.style.color = 'green';
    dcButton.style.visibility = 'visible'
    console.log('SSE connection established');
  };
  
  eventSource.onerror = function(error) {
    messageDiv.textContent = 'Connection error. Please try again.';
    messageDiv.style.color = 'red';
    dcButton.style.visibility = 'hidden'
    pingDiv.textContent = ``;
    console.error('SSE connection error:', error);
    eventSource.close();
  };

  eventSource.addEventListener("close", function(event) {
    console.log("Server is shutting down"); //todo add event data to this
    eventSource.close();

    messageDiv.textContent = 'Server shutting down... Disconnected';
    messageDiv.style.color = 'blue';
    dcButton.style.visibility = 'hidden'
    pingDiv.textContent = ``;
  });
  
  eventSource.addEventListener("ping", function(event) {
    const serverTimestamp = parseInt(event.data, 10);
    console.log(`Ping: server timestamp: ${serverTimestamp}`);
    /*pingDiv.textContent = `Ping: ${ping}ms`;
    pingDiv.style.color = '#064a72';*/
  });
}

async function loginAndConnect() {
  const username = document.getElementById('username').value;
  const key = document.getElementById('key').value;
  
  messageDiv.textContent = 'Authenticating...';
  messageDiv.style.color = 'blue';
  
  const success = await authenticate(username, key);
  
  if (success) {
    messageDiv.textContent = 'Authentication successful. Connecting...';
    connectSSE();
  } else {
    messageDiv.textContent = 'Authentication failed. Please check your credentials.';
    messageDiv.style.color = 'red';
  }
}

document.getElementById('authForm').addEventListener('submit', function(e) {
  e.preventDefault();
  loginAndConnect();
});

dcButton.addEventListener('click', function() {
    if (eventSource) {
        eventSource.close();
    }
    messageDiv.textContent = 'Disconnected';
    messageDiv.style.color = 'blue';
    dcButton.style.visibility = 'hidden'
    pingDiv.textContent = ``;
});

// Cleanup function for page unload
window.addEventListener('beforeunload', function() {
  if (eventSource) {
    eventSource.close();
  }
});
