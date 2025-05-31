// Main application logic
function updateUI(data) {
    // CPU information
    const cpuUsage = data.current_metrics.cpus.reduce((acc, cpu) => acc + cpu.usage, 0) / data.current_metrics.cpus.length;
    document.getElementById('cpu-usage').textContent = `${cpuUsage.toFixed(1)}%`;
    document.getElementById('cpu-progress').style.width = `${cpuUsage}%`;
    
    // CPU cores details
    document.getElementById('cpu-cores-list').innerHTML = data.current_metrics.cpus
        .map((cpu, index) => `
            <div class="mb-2">
                <div class="d-flex justify-content-between">
                    <small>Core ${index}</small>
                    <small>${cpu.usage.toFixed(1)}% @ ${cpu.frequency}MHz</small>
                </div>
                <div class="progress" style="height: 4px;">
                    <div class="progress-bar" role="progressbar" 
                         style="width: ${cpu.usage}%" 
                         aria-valuenow="${cpu.usage}" 
                         aria-valuemin="0" 
                         aria-valuemax="100"></div>
                </div>
            </div>
        `).join('');
    
    // Memory information
    const memoryPercent = (data.current_metrics.memory.used / data.current_metrics.memory.total) * 100;
    document.getElementById('memory-usage').textContent = `${memoryPercent.toFixed(1)}%`;
    document.getElementById('memory-progress').style.width = `${memoryPercent}%`;
    
    document.getElementById('memory-detail').innerHTML = `
        <div class="d-flex justify-content-between">
            <small>Total:</small>
            <small>${formatBytes(data.current_metrics.memory.total * 1024)}</small>
        </div>
        <div class="d-flex justify-content-between">
            <small>Used:</small>
            <small>${formatBytes(data.current_metrics.memory.used * 1024)}</small>
        </div>
        <div class="d-flex justify-content-between">
            <small>Available:</small>
            <small>${formatBytes(data.current_metrics.memory.available * 1024)}</small>
        </div>
    `;
    
    // Disk information
    document.getElementById('disk-list').innerHTML = data.current_metrics.disks
        .map(disk => `
            <div class="mb-3">
                <div class="d-flex justify-content-between mb-1">
                    <span>${disk.name}</span>
                    <span>${disk.percent_used.toFixed(1)}% used</span>
                </div>
                <div class="progress">
                    <div class="progress-bar ${disk.percent_used < 70 ? 'bg-success' : (disk.percent_used < 90 ? 'bg-warning' : 'bg-danger')}" 
                         role="progressbar" style="width: ${disk.percent_used}%"></div>
                </div>
                <div class="d-flex justify-content-between mt-1">
                    <small>${formatBytes(disk.available_space * 1024)} free</small>
                    <small>${formatBytes(disk.total_space * 1024)} total</small>
                </div>
            </div>`
        ).join('');
    
    // System information
    document.getElementById('hostname').textContent = data.current_metrics.hostname;
    document.getElementById('uptime').textContent = formatUptime(data.current_metrics.uptime);
    
    // Anomaly detection
    const cpuAnomalyBadge = document.getElementById('cpu-anomaly-badge');
    const memoryAnomalyBadge = document.getElementById('memory-anomaly-badge');
    const cpuCard = document.getElementById('cpu-card');
    const memoryCard = document.getElementById('memory-card');
    
    if (data.anomalies.cpu_anomaly) {
        cpuAnomalyBadge.className = 'badge bg-danger';
        cpuAnomalyBadge.textContent = 'ANOMALY DETECTED';
        cpuCard.classList.add('anomaly-alert');
    } else {
        cpuAnomalyBadge.className = 'badge bg-success';
        cpuAnomalyBadge.textContent = 'Normal';
        cpuCard.classList.remove('anomaly-alert');
    }
    
    if (data.anomalies.memory_anomaly) {
        memoryAnomalyBadge.className = 'badge bg-danger';
        memoryAnomalyBadge.textContent = 'ANOMALY DETECTED';
        memoryCard.classList.add('anomaly-alert');
    } else {
        memoryAnomalyBadge.className = 'badge bg-success';
        memoryAnomalyBadge.textContent = 'Normal';
        memoryCard.classList.remove('anomaly-alert');
    }
    
    // Update timestamp
    const date = new Date(data.timestamp * 1000);
    document.getElementById('refresh-badge').textContent = 
        `Last updated: ${date.toLocaleTimeString()}`;
    
    // Flash the refresh badge
    const badge = document.getElementById('refresh-badge');
    badge.style.opacity = '0.5';
    setTimeout(() => {
        badge.style.opacity = '1';
    }, 200);
}

function fetchHistory() {
    fetch(HISTORY_ENDPOINT)
        .then(response => {
            if (!response.ok) throw new Error('Network response was not ok');
            return response.json();
        })
        .then(data => {
            cpuData = data.cpu;
            memoryData = data.memory;
            const labels = Array(data.cpu.length).fill('');
            
            cpuChart.data.labels = labels;
            cpuChart.data.datasets[0].data = cpuData;
            cpuChart.update();
            
            memoryChart.data.labels = labels;
            memoryChart.data.datasets[0].data = memoryData;
            memoryChart.update();
        })
        .catch(error => console.error('Error fetching history:', error));
}

function fetchDashboard() {
    document.getElementById('error-alert').classList.add('d-none');
    
    fetch(API_ENDPOINT)
        .then(response => {
            if (!response.ok) throw new Error('Network response was not ok');
            return response.json();
        })
        .then(data => {
            updateUI(data);
            fetchHistory();
        })
        .catch(error => {
            console.error('Error fetching dashboard data:', error);
            document.getElementById('error-alert').classList.remove('d-none');
        });
}

function setupRefresh() {
    fetchDashboard(); // Initial fetch
    setInterval(fetchDashboard, REFRESH_INTERVAL);
    document.getElementById('refresh-btn').addEventListener('click', fetchDashboard);
}

// Initialize everything when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    initCharts();
    setupRefresh();
}); 