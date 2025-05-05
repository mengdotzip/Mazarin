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
  
  eventSource = new EventSource('/sse');
  
  eventSource.onopen = function() {
    messageDiv.textContent = 'Successfully authenticated!';
    messageDiv.style.color = 'green';
    dcButton.style.visibility = 'visible'
    console.log('SSE connection established');
  };
  
  eventSource.onerror = function(error) {
    messageDiv.textContent = 'Connection error. Please try again.';
    messageDiv.style.color = 'red';
    dcButton.style.visibility = 'hidden'
    console.error('SSE connection error:', error);
    eventSource.close();
  };

  eventSource.addEventListener("close", function(event) {
    console.log("Server is shutting down"); //todo add event data to this
    eventSource.close();

    messageDiv.textContent = 'Server shutting down... Disconnected';
    messageDiv.style.color = 'blue';
    dcButton.style.visibility = 'hidden'
  });
  
  eventSource.addEventListener("ping", function(event) {
    const now = Date.now();
    const serverTimestamp = parseInt(event.data, 10);
    const ping = now - serverTimestamp;
    console.log(`Ping: ${ping}ms`);
    pingDiv.textContent = `Ping: ${ping}ms`;
    pingDiv.style.color = '#064a72';
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
});

// Cleanup function for page unload
window.addEventListener('beforeunload', function() {
  if (eventSource) {
    eventSource.close();
  }
});
