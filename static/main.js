


if (socket){
    socket.close()
}

    var socket = new WebSocket("/admin/ws/main_graph")

    var ctx = document.getElementById('myChart').getContext('2d');
    var myChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: [], 
            datasets: [
                {
                    label: 'Requests per second',
                    data: [],
                    backgroundColor: 'rgba(54, 162, 235, 0.2)',
                    borderColor: '#9F9FF8',
                    borderWidth: 1,
                    fill: false,
                   
                    yAxisID: 'y' 
                },
                {
                    label: 'Memory usage (MB)',
                    data: [], 
                    backgroundColor: 'rgba(255, 99, 132, 0.2)',
                    borderColor: 'rgba(255, 99, 132, 1)', 
                    borderWidth: 1,
                    fill: false,
                   
                    yAxisID: 'y1'
                }
            ]
        },
        options: {
            responsive: true,
            scales: {
                x: {
                    type: 'time', 
                    time: {
                        unit: 'second', 
                        stepSize:5,
                        tooltipFormat: 'PPpp', 
                        displayFormats: {
                            second: 'HH:mm:ss' 
                        }
                    },
                    title: {
                        display: true,
                        text: 'Time',
                        color: '#9F9FF8' 
                    },
                    ticks: {
                        source: 'auto', 
                        autoSkip: false,
                        maxRotation: 0, 
                        color: 'rgba(255, 255, 255, 0.40)'
                    }
                },
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Requests',
                        color: '#9F9FF8'
                    },
                    ticks: {
                        color: 'rgba(255, 255, 255, 0.40)' 
                    },
                    suggestedMin: 0, 
                    suggestedMax: 100 
                },
                y1: {
                    beginAtZero: true,
                    position: 'right',
                    title: {
                        display: true,
                        text: 'Memory (MB)',
                        color: '#92BFFF' 
                    },
                    ticks: {
                        color: 'rgba(255, 255, 255, 0.40)' 
                    },
                    grid: {
                        drawOnChartArea: false 
                    },
                    suggestedMin: 50, 
                    suggestedMax: 150 
                }
            },
            plugins: {
                legend: {
                    labels: {
                        color: '#FFFFFF' 
                    }
                }
            }
        }
    });

   

   

    var allData = {
        labels: [],
        rps: [],
        mem: []
    };

    
    

   
    socket.onmessage = function(event) {
        let sample = JSON.parse(event.data);

      
        allData.labels.push(new Date());
        allData.rps.push(sample.rps);
        allData.mem.push(sample.mem);

        
        
        updateChart();
    };

   
    function updateChart() {
       
        console.log(allData);

        console.log( allData.labels.slice(-8))

        myChart.data.labels = allData.labels.slice(-10);
        myChart.data.datasets[0].data = allData.rps.slice(-10);
        myChart.data.datasets[1].data = allData.mem.slice(-10);

        myChart.update();
    }

