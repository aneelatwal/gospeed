document.getElementById("startBtn").addEventListener("click", async () => {
  const output = document.getElementById("output");
  output.innerHTML = "Running test... please wait.";

  try {
    const res = await fetch("/api/speedtest");
    const data = await res.json();
    output.innerHTML = `
      <h4>Results</h4>
      <p><strong>Server:</strong> ${data.server}</p>
      <p><strong>Download:</strong> ${data.download_mbps.toFixed(2)} Mbps</p>
      <p><strong>Upload:</strong> ${data.upload_mbps.toFixed(2)} Mbps</p>
      <p><small>Timestamp: ${data.timestamp}</small></p>
    `;
  } catch (err) {
    output.innerHTML = "Error running test: " + err.message;
  }
});