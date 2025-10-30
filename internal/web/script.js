const startBtn = document.getElementById("startBtn");
const output = document.getElementById("output");
const spinner = document.getElementById("spinner");

async function fetchAndRenderHistory() {
  const container = document.getElementById("history");
  container.innerHTML = "<div class='text-gray-400 mb-2 text-lg font-semibold'>Recent Tests</div>";
  try {
    const res = await fetch("/api/history");
    const history = await res.json();
    if (!Array.isArray(history) || history.length === 0) {
      container.innerHTML += "<div class='text-gray-500'>No history yet.</div>";
      return;
    }
    // Render newest first
    const rows = history.reverse().map(r => {
      const d = new Date(r.Timestamp || r.timestamp);
      const dateStr = d.toLocaleString([], {dateStyle:"medium", timeStyle:"short"});
      return `<tr>
        <td class='px-2 py-1'>${dateStr}</td>
        <td class='px-2 py-1 text-right text-green-400'>${Number(r.PingMs ?? r.ping_ms).toFixed(0)} ms</td>
        <td class='px-2 py-1 text-right text-green-400'>${Number(r.DownloadMbps ?? r.download_mbps).toFixed(2)} Mbps</td>
        <td class='px-2 py-1 text-right text-green-400'>${Number(r.UploadMbps ?? r.upload_mbps).toFixed(2)} Mbps</td>
      </tr>`;
    }).join("");
    container.innerHTML += `
      <table class='w-full text-sm text-gray-100 bg-[#171c27] rounded-lg'>
        <thead><tr class='border-b border-gray-700'>
          <th class='px-2 py-1 text-left'>Time</th>
          <th class='px-2 py-1 text-right'>Ping</th>
          <th class='px-2 py-1 text-right'>Download</th>
          <th class='px-2 py-1 text-right'>Upload</th>
        </tr></thead>
        <tbody>${rows}</tbody>
      </table>`;
  } catch(e) {
    container.innerHTML += `<div class='text-red-400'>Error loading history</div>`;
  }
}

window.addEventListener("DOMContentLoaded", fetchAndRenderHistory);

startBtn.addEventListener("click", async () => {
  // Add pulsing animation and show spinner
  startBtn.classList.add("animate-pulse", "opacity-80");
  spinner.classList.remove("hidden");
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

    fetchAndRenderHistory(); // Refresh results after speedtest

  } catch (err) {
    document.getElementById("timestamp").textContent = "Error: " + err.message;
  } finally {
    // Stop pulsing and hide spinner
    startBtn.classList.remove("animate-pulse", "opacity-80");
    spinner.classList.add("hidden");
  }
});