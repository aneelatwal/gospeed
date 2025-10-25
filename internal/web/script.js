const startBtn = document.getElementById("startBtn");
const output = document.getElementById("output");

startBtn.addEventListener("click", async () => {
  // Add pulsing animation
  startBtn.classList.add("animate-pulse", "opacity-80");
  output.classList.remove("hidden");

  document.getElementById("pingVal").textContent = "--";
  document.getElementById("downVal").textContent = "--";
  document.getElementById("upVal").textContent = "--";
  document.getElementById("timestamp").textContent = "Running...";

  try {
    const res = await fetch("/api/speedtest");
    const data = await res.json();

    document.getElementById("pingVal").textContent = data.ping_ms.toFixed(0);
    document.getElementById("downVal").textContent = data.download_mbps.toFixed(2);
    document.getElementById("upVal").textContent = data.upload_mbps.toFixed(2);
    document.getElementById("timestamp").textContent = new Date(data.timestamp)
      .toLocaleString([], { dateStyle: "medium", timeStyle: "short" });

  } catch (err) {
    document.getElementById("timestamp").textContent = "Error: " + err.message;
  } finally {
    // Stop pulsing
    startBtn.classList.remove("animate-pulse", "opacity-80");
  }
});