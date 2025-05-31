// Chart initialization and configuration
let cpuChart, memoryChart;
let cpuData = [], memoryData = [];

function initCharts() {
    const chartConfig = {
        type: 'line',
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true,
                    max: 100,
                    title: {
                        display: true,
                        text: 'Usage %'
                    }
                }
            },
            animation: false
        }
    };

    // CPU Chart
    cpuChart = new Chart(document.getElementById('cpu-chart'), {
        ...chartConfig,
        data: {
            labels: [],
            datasets: [{
                label: 'CPU Usage %',
                data: [],
                borderColor: 'rgb(75, 192, 192)',
                tension: 0.1
            }]
        }
    });

    // Memory Chart
    memoryChart = new Chart(document.getElementById('memory-chart'), {
        ...chartConfig,
        data: {
            labels: [],
            datasets: [{
                label: 'Memory Usage %',
                data: [],
                borderColor: 'rgb(153, 102, 255)',
                tension: 0.1
            }]
        }
    });
} 